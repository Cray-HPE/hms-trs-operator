@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-trs-operator"
        repository = "cray"
        imagePrefix = "hms"
        app = "trs-operator"
        name = "hms-trs-operator"
        description = "Cray HMS TRS Operator"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
