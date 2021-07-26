/*
 * MIT License
 *
 * (C) Copyright [2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package trsworker

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sort"
	trs_kafka "github.com/Cray-HPE/hms-trs-kafkalib/pkg/trs-kafkalib"
	"github.com/Cray-HPE/hms-trs-operator/pkg/apis/kafka/v1beta1"
	trsv1alpha1 "github.com/Cray-HPE/hms-trs-operator/pkg/apis/trs/v1alpha1"
	"github.com/Cray-HPE/hms-trs-operator/pkg/kafka_topics"
	"strings"
)

const WorkerCount = 3
const TRSWokerAppLabel = "trsworker"

var log = logf.Log.WithName("controller_trsworker")

// Add creates a new TRSWorker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTRSWorker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("trsworker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TRSWorker
	err = c.Watch(&source.Kind{Type: &trsv1alpha1.TRSWorker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TRSWorker
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &trsv1alpha1.TRSWorker{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTRSWorker implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTRSWorker{}

// ReconcileTRSWorker reconciles a TRSWorker object
type ReconcileTRSWorker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TRSWorker object and makes changes based on the state read
// and what is in the TRSWorker.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTRSWorker) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TRSWorker")

	// In all honesty I tried this both ways and didn't see a difference, but every example I found would requeue
	// after doing every CRUD action...so that's what I'm defaulting to here. By the very end of this function if we
	// haven't done any consequential work, we'll set this to false and return.
	result = reconcile.Result{
		Requeue: true,
	}

	// Fetch the TRSWorker resource. This is specific to each service.
	trsWorker := &trsv1alpha1.TRSWorker{}
	serviceGetErr := r.client.Get(context.TODO(), request.NamespacedName, trsWorker)

	if serviceGetErr != nil {
		// Any error except for not found is considered a failure.
		if errors.IsNotFound(serviceGetErr) {
			// Request object not found, must have been deleted.
			// Check all the TRS deployments to see if any are now zombies and need to be reaped.
			trsDeployments, err := r.getTRSDeployments(request.Namespace)
			if err != nil {
				reqLogger.Error(err, "Unable to get TRS deployments")
				resultErr = err
				return
			}

			missingLabelErr := fmt.Errorf("deployment missing label")

			// Loop through all of the TRS deployments and check for TRS workers using that deployment.
			for _, deployment := range trsDeployments {
				// Check to make sure the labels we need are there. To prevent the operator from bombing out over and
				// over because of a missing label, just skip past ones that don't have them. In truth this is
				// something that shouldn't ever happen, but given that returning an error from reconcile just
				// results in a retried attempt over and over, that's not going to help anything.
				workerType, ok := deployment.Spec.Selector.MatchLabels["worker_type"]
				if !ok {
					reqLogger.Error(missingLabelErr, "Found TRS deployment, but missing worker_type label",
						"Deployment.Name", deployment.Name)
					continue
				}
				workerVersion, ok := deployment.Spec.Selector.MatchLabels["worker_version"]
				if !ok {
					reqLogger.Error(missingLabelErr, "Found TRS deployment, but missing worker_version label",
						"Deployment.Name", deployment.Name)
					continue
				}

				// For each deployment, get all the TRSWorker deployments matching the spec.
				thisSpec := trsv1alpha1.TRSWorkerSpec{
					WorkerType:    workerType,
					WorkerVersion: workerVersion,
				}

				workersMatchingSpec, getWorkersErr := r.getTRSWorkersForSpec(thisSpec, trsWorker.Namespace)
				if getWorkersErr != nil {
					reqLogger.Error(getWorkersErr, "Failed to get workers for spec",
						"TRSWorkerSpec", thisSpec)
					resultErr = getWorkersErr
					return
				}

				// If the length of the array of workers using this spec is 0, then this deployment is a zombie and
				// needs to be reaped. TODO: This also means the ConfigMap for this deployment can also be removed.
				if len(workersMatchingSpec) == 0 {
					deleteLogger := reqLogger.WithValues("Deployment.Name", deployment.Name)

					deleteErr := r.client.Delete(context.TODO(), &deployment)
					if deleteErr != nil {
						deleteLogger.Error(deleteErr, "Failed to delete zombie deployment")
						resultErr = deleteErr
						return
					}

					deleteLogger.Info("Deleted deployment as no remaining TRS workers need it anymore")
				} else {
					var workerNames []string
					for _, worker := range workersMatchingSpec {
						workerNames = append(workerNames, worker.Name)
					}

					reqLogger.Info("Deployment still has dependent workers, not removing",
						"Deployment.Name", deployment.Name, "TRSWorkers.Name", workerNames)

					// Now we need to update the ConfigMap for this deployment to reflect this TRSWorker being removed.
					configMap, configMapUpdated, configMapErr := r.createOrUpdateDeploymentConfigMap(&deployment,
						thisSpec)
					if configMapErr != nil {
						reqLogger.Error(configMapErr, "Failed to generate ConfigMap for Deployment",
							"Deployment.Name", deployment.Name)
						resultErr = configMapErr
						return
					} else if configMapUpdated {
						reqLogger.Info("Updated ConfigMap for Deployment",
							"ConfigMap.Name", configMap.Name,
							"ConfigMap.Namespace", configMap.Namespace,
							"ConfigMap.Data", configMap.Data)
					}
				}
			}

			reqLogger.Info("TRS worker removed")
			result.Requeue = false
			return
		} else {
			// Error reading the object - requeue the request.
			reqLogger.Error(serviceGetErr, "Failed to get TRSWorker")
			resultErr = serviceGetErr
			return
		}
	} else {
		// In this block are all the activities that should only be done on valid TRSWorker objects (i.e., ones that
		// haven't been removed).

		// Check to see if the necessary KafkaTopic CRs exist.
		// Start by generating them so we know what names to look for.
		workerKafkaTopics := r.generateKafkaTopicsForTRSWorker(trsWorker)
		createdTopic := false
		// Now for each attempt to find it. If it doesn't exist, create it.
		for _, topic := range workerKafkaTopics {
			foundTopic := &v1beta1.KafkaTopic{}
			findErr := r.client.Get(context.TODO(), types.NamespacedName{
				Name:      topic.Name,
				Namespace: topic.Namespace,
			}, foundTopic)

			kafkaLogger := reqLogger.WithValues("KafkaTopic.Namespace", topic.Namespace,
				"KafkaTopic.Name", topic.Name)

			if findErr != nil && errors.IsNotFound(findErr) {
				createErr := r.client.Create(context.TODO(), topic)
				if createErr != nil {
					kafkaLogger.Error(createErr, "Failed to create new KafkaTopic")
					resultErr = createErr
					return
				}

				kafkaLogger.Info("Created KafkaTopic")
				createdTopic = true
			} else if findErr != nil {
				reqLogger.Error(findErr, "Failed to get KafkaTopic")
				resultErr = findErr
				return
			} else {
				// Make sure the status of the latest generation of this topic is Ready.
				if len(foundTopic.Status.Conditions) == 0 ||
					foundTopic.Status.Conditions[len(foundTopic.Status.Conditions)-1].Type != "Ready" {
					kafkaLogger.Info("Topic not ready yet")
					return
				}
			}
		}
		if createdTopic {
			// Don't requeue after each topic creation, that's wasteful.
			return
		}
	}

	/*
	 * If we've made it to here it's most likely because a new TRS Worker CR has been created, though it's important
	 * to remember that's not necessarily true so we need to check where applicable! The only thing we know for sure
	 * at this point is this specific TRSWorker exists.
	 *
	 * Our mission at this point is to make sure the proper configuration exists for _all_ other parts of TRS.
	 *
	 * From here on out *ONLY* those activities that are shared between TRSWorkers should be considered. If something
	 * is worker specific, back up the truck, you will get in trouble putting it below this line.
	 */

	// Check to see if a new deployment is necessary.
	// Remember, with TRS deployments of the same type and version are shared, so this only happens if one doesn't
	// already exist.

	// Each deployment name has the format trsworker-WORKER_TYPE-WORKER_VERSION, and must be all lowercase.
	deploymentName := strings.ToLower(strings.Join([]string{
		"cray-hms",
		TRSWokerAppLabel,
		trsWorker.Spec.WorkerType,
		trsWorker.Spec.WorkerVersion,
	}, "-"))

	foundDeployment := &v1.Deployment{}
	deploymentGetErr := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: trsWorker.Namespace,
		Name:      deploymentName,
	}, foundDeployment)

	if deploymentGetErr != nil && errors.IsNotFound(deploymentGetErr) {
		// Deployment for this worker type and version doesn't exist, make one.
		newDeployment := r.generateDeploymentForTRSWorker(trsWorker, deploymentName)
		deploymentLogger := reqLogger.WithValues("Deployment.Namespace", newDeployment.Namespace,
			"Deployment.Name", newDeployment.Name,
			"TRSWorker.Spec.WorkerType", trsWorker.Spec.WorkerType,
			"TRSWorker.Spec.WorkerVersion", trsWorker.Spec.WorkerVersion)

		createErr := r.client.Create(context.TODO(), newDeployment)
		if createErr != nil {
			deploymentLogger.Error(createErr, "Failed to create Deployment")
			resultErr = createErr
			return
		}

		deploymentLogger.Info("Deployment created")
		return
	} else if deploymentGetErr != nil {
		reqLogger.Error(deploymentGetErr, "Failed to get Deployment")
		resultErr = deploymentGetErr
		return
	}

	// Check to make sure the environment variables are the way they should be. Start by finding the right container.
	// Technically, there's only the one, but this is safe and futureproof against sidecars.
	for containerIndex, _ := range foundDeployment.Spec.Template.Spec.Containers {
		container := &foundDeployment.Spec.Template.Spec.Containers[containerIndex]
		if container.Name == TRSWokerAppLabel {
			correctEnv := environmentVariablesForDeployment()
			if !reflect.DeepEqual(container.Env, correctEnv) {
				containerLogger := reqLogger.WithValues("Deployment.Namespace", foundDeployment.Namespace,
					"Deployment.Name", foundDeployment.Name,
					"Container.Name", container.Name,
					"Container.Env-old", container.Env,
					"Container.Env-new", correctEnv)
				container.Env = correctEnv

				updateErr := r.client.Update(context.TODO(), foundDeployment)
				if updateErr != nil {
					containerLogger.Error(updateErr, "Failed to update Deployment with new Env")
					resultErr = updateErr
					return
				}

				containerLogger.Info("Updated Deployment with new Env")
				return
			}
		}
	}

	// Now that the deployment is the way it should be we need to check the ConfigMap for this deployment to make sure
	// it exists and has everything it needs.
	configMap, configMapUpdated, configMapErr := r.createOrUpdateDeploymentConfigMap(foundDeployment, trsWorker.Spec)
	if configMapErr != nil {
		reqLogger.Error(configMapErr, "Failed to generate ConfigMap for Deployment",
			"Deployment.Name", foundDeployment.Name)
	} else if configMapUpdated {
		reqLogger.Info("Created/updated ConfigMap for Deployment",
			"ConfigMap.Name", configMap.Name,
			"ConfigMap.Namespace", configMap.Namespace,
			"ConfigMap.Data", configMap.Data)
		return
	}

	reqLogger.Info("Reconcile skipped, configuration already meets needs")
	result.Requeue = false
	return
}

// Creates or updates a ConfigMap for a given Deployment and spec.
func (r *ReconcileTRSWorker) createOrUpdateDeploymentConfigMap(deployment *v1.Deployment,
	spec trsv1alpha1.TRSWorkerSpec) (*corev1.ConfigMap, bool, error) {
	deploymentConfigMap, configMapErr := r.generateConfigMapForDeployment(deployment, spec)
	if configMapErr != nil {
		return nil, false, configMapErr
	}

	foundConfigMap := &corev1.ConfigMap{}
	configMapGetErr := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deploymentConfigMap.Name,
	}, foundConfigMap)
	if configMapGetErr != nil && errors.IsNotFound(configMapGetErr) {
		// ConfigMap for this Deployment doesn't exist, make one.
		createErr := r.client.Create(context.TODO(), deploymentConfigMap)
		if createErr != nil {
			return nil, false, createErr
		}

		return deploymentConfigMap, true, nil
	} else if configMapGetErr != nil {
		return nil, false, configMapGetErr
	}

	// ConfigMap already exists, check the contents to ensure they are good.
	if !reflect.DeepEqual(foundConfigMap.Data, deploymentConfigMap.Data) {
		foundConfigMap.Data = deploymentConfigMap.Data

		updateErr := r.client.Update(context.TODO(), foundConfigMap)
		if updateErr != nil {
			return nil, false, updateErr
		}

		return foundConfigMap, true, nil
	}

	return foundConfigMap, false, nil
}

// Returns a list of Kafka topics a given spec pertains to. This is computed by looking at all the active TRSWorker
// configurations and cross referencing their configuration with this spec to either include them or not.
func (r *ReconcileTRSWorker) getKafkaTopicsForSpec(spec trsv1alpha1.TRSWorkerSpec, namespace string) (topics []string,
	err error) {
	// Start by getting the appropriate TRS workers.
	trsWorkers, getWorkersErr := r.getTRSWorkersForSpec(spec, namespace)
	if getWorkersErr != nil {
		err = fmt.Errorf("unable to generate Kafka topics because get of workers failed: %w", getWorkersErr)
		return
	}

	// All will have the same workerClass.
	workerClass := fmt.Sprintf("%s-%s", spec.WorkerType, spec.WorkerVersion)
	for _, worker := range trsWorkers {
		// The only topic that matters here is the one the client will be sending on (and therefore the worker
		// needs to be listening on)
		sendTopicName, _, _ := trs_kafka.GenerateSendReceiveConsumerGroupName(worker.Name,
			workerClass, "")
		topics = append(topics, sendTopicName)
	}

	return
}

// Returns a new KafkaTopic object for a TRS worker.
func (r *ReconcileTRSWorker) generateKafkaTopicsForTRSWorker(trsWorker *trsv1alpha1.TRSWorker) []*v1beta1.KafkaTopic {
	kafkaClusterName, found := os.LookupEnv("TRS_KAFKA_CLUSTER_NAME")
	if !found {
		kafkaClusterName = "cray-shared-kafka"
	}

	// Generate the topic names using the core TRS Kafka library.
	sendTopicName, receiveTopicName, _ := trs_kafka.GenerateSendReceiveConsumerGroupName(trsWorker.Name,
		fmt.Sprintf("%s-%s", trsWorker.Spec.WorkerType, trsWorker.Spec.WorkerVersion), "")

	// Keep the configuration the same for both.
	topicSpec := v1beta1.KafkaTopicSpec{
		Config: map[string]string{
			// These settings are the same as what are used for the telemetry topics in SMF. They're overkill.
			"retention.ms":           "14400000",
			"segment.bytes":          "1048576",
		},
		// Make the number of partitions twice the number of workers so that if we increase size later on the
		// rebalancing will be easy.
		Partitions: WorkerCount * 2,
		Replicas:   2,
	}

	// Both get the same labels as well.
	labels := map[string]string{
		"strimzi.io/cluster": kafkaClusterName,
	}

	// Make two topics, one for the send and one for the receive.
	sendTopic := &v1beta1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sendTopicName,
			Namespace: trsWorker.Namespace,
			Labels:    labels,
		},
		Spec: topicSpec,
	}
	sendTopic.Spec.TopicName = sendTopicName

	receiveTopic := &v1beta1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      receiveTopicName,
			Namespace: trsWorker.Namespace,
			Labels:    labels,
		},
		Spec: topicSpec,
	}
	receiveTopic.Spec.TopicName = receiveTopicName

	// Set TRSWorker instance as the owner and controller. This will allow the topics to be cleaned up automatically.
	_ = controllerutil.SetControllerReference(trsWorker, sendTopic, r.scheme)
	_ = controllerutil.SetControllerReference(trsWorker, receiveTopic, r.scheme)

	return []*v1beta1.KafkaTopic{sendTopic, receiveTopic}
}

// Returns a ConfigMap object for a worker type and version.
func (r *ReconcileTRSWorker) generateConfigMapForDeployment(deployment *v1.Deployment,
	spec trsv1alpha1.TRSWorkerSpec) (configMap *corev1.ConfigMap, err error) {
	// Start by getting the contents of the Kafka topics block.
	topics, topicsErr := r.getKafkaTopicsForSpec(spec, deployment.Namespace)
	if topicsErr != nil {
		err = fmt.Errorf("unable to generate ConfigMap because of topic err: %w", topicsErr)
		return
	}

	// Sort the topics so we get the same list every time. Otherwise this method can flap back and forth.
	sort.Strings(topics)

	// Make up a structure to put the topics in.
	var kafkaConfig []kafka_topics.KafkaTopic
	for _, topic := range topics {
		newTopicConfig := kafka_topics.KafkaTopic{
			TopicName: topic,
		}
		kafkaConfig = append(kafkaConfig, newTopicConfig)
	}

	// Now make a (pretty) string out of the config.
	jsonBytes, jsonErr := json.MarshalIndent(kafkaConfig, "", "  ")
	if jsonErr != nil {
		err = fmt.Errorf("unable to generate ConfigMap because of JSON err: %w", jsonErr)
		return
	}

	configMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{deployment.Name, "config"}, "-"),
			Namespace: deployment.Namespace,
		},
		Data: map[string]string{
			"active_topics.json": string(jsonBytes),
		},
	}

	// Set Deployment as the owner and controller. This will allow the ConfigMap to be cleaned up automatically.
	_ = controllerutil.SetControllerReference(deployment, configMap, r.scheme)

	return
}

// Returns a TRSWorker Deployment object for a worker type and version.
func (r *ReconcileTRSWorker) generateDeploymentForTRSWorker(trsWorker *trsv1alpha1.TRSWorker,
	deploymentName string) *v1.Deployment {
	ls := labelsForSpec(trsWorker.Spec)
	// TODO: Remove hardcoded size...maybe?
	var replicas int32
	replicas = WorkerCount

	// These are environment variables that the worker deployment pods need that we pass through the operator.
	imagePrefix, found := os.LookupEnv("TRS_IMAGE_PREFIX")
	if !found {
		imagePrefix = ""
	}

	// Compute the image name from the TRSWorker spec.
	imageName := strings.Join([]string{"hms", "trs", "worker",
		trsWorker.Spec.WorkerType, trsWorker.Spec.WorkerVersion}, "-")
	fullImage := imagePrefix + "cray/" + imageName

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: trsWorker.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string {
						"traffic.sidecar.istio.io/excludeOutboundPorts": "8082,9092,2181",
					},
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  TRSWokerAppLabel,
						Image: fullImage,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: "/configs",
							},
						},
						Env: environmentVariablesForDeployment(),
					}},
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: strings.Join([]string{deploymentName, "config"}, "-"),
									},
								}},
						},
					},
				},
			},
		},
	}

	return deployment
}

// Gets all the deployments matching the app label.
func (r *ReconcileTRSWorker) getTRSDeployments(namespace string) (deployments []v1.Deployment, err error) {
	deploymentsList := &v1.DeploymentList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}

	err = r.client.List(context.TODO(), deploymentsList, listOpts...)
	if err != nil {
		return
	}

	for _, deployment := range deploymentsList.Items {
		if deployment.Spec.Selector.MatchLabels["app"] == TRSWokerAppLabel {
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}

// Returns the list of Deployments for a given worker class and version.
func (r *ReconcileTRSWorker) getDeploymentsForSpec(spec trsv1alpha1.TRSWorkerSpec,
	namespace string) (deployments []v1.Deployment, err error) {
	trsDeployments, err := r.getTRSDeployments(namespace)
	if err != nil {
		return
	}

	for _, deployment := range trsDeployments {
		if deployment.Spec.Selector.MatchLabels["worker_type"] == spec.WorkerType &&
			deployment.Spec.Selector.MatchLabels["worker_version"] == spec.WorkerVersion {
			deployments = append(deployments, deployment)
		}
	}

	return
}

// Returns the list of TRSWorkers for a given worker class and version.
func (r *ReconcileTRSWorker) getTRSWorkersForSpec(spec trsv1alpha1.TRSWorkerSpec,
	namespace string) (workers []trsv1alpha1.TRSWorker, err error) {
	trsWorkerList := &trsv1alpha1.TRSWorkerList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}

	err = r.client.List(context.TODO(), trsWorkerList, listOpts...)
	if err != nil {
		return
	}

	// Loop through the returned TRSWorkers checking to see if their type and version match. If so, add them.
	for _, trsWorker := range trsWorkerList.Items {
		if trsWorker.Spec.WorkerType == spec.WorkerType && trsWorker.Spec.WorkerVersion == spec.WorkerVersion {
			workers = append(workers, trsWorker)
		}
	}

	return
}

// Returns the labels for selecting the resources belonging to the given memcached CR name.
func labelsForSpec(spec trsv1alpha1.TRSWorkerSpec) map[string]string {
	return map[string]string{
		"app":            TRSWokerAppLabel,
		"worker_type":    spec.WorkerType,
		"worker_version": spec.WorkerVersion,
	}
}

// Returns environment variables for a deployment
func environmentVariablesForDeployment() []corev1.EnvVar {
	workerLogLevel, found := os.LookupEnv("TRS_WORKER_LOG_LEVEL")
	if !found {
		workerLogLevel = "INFO"
	}
	workerKafkaBrokerSpec, found := os.LookupEnv("TRS_WORKER_KAFKA_BROKER_SPEC")
	if !found {
		workerKafkaBrokerSpec = "kafka:9092"
	}

	env := []corev1.EnvVar{
		{
			Name:  "LOG_LEVEL",
			Value: workerLogLevel,
		},
		{
			Name:  "BROKER_SPEC",
			Value: workerKafkaBrokerSpec,
		},
	}

	return env
}
