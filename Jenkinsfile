@Library('dst-shared@release/shasta-1.4') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "hms"
        app = "trs-operator"
        name = "hms-trs-operator"
        description = "Cray HMS TRS Operator"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
