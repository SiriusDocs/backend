pipeline {
    agent any
    
    environment {
        GIT_CREDENTIALS_ID = 'git-credentials'
        REGISTRY_CREDENTIALS_ID = 'registry-credentials'
        REGISTRY_URL = 'registry.certsirius.ru'
        IMAGE_TAG = "${env.BUILD_NUMBER}"
        NAMESPACE = "siriusdocs"
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: '*/main']],
                    userRemoteConfigs: [[
                        url: 'https://github.com/SiriusDocs/backend.git',
                        credentialsId: env.GIT_CREDENTIALS_ID
                    ]]
                ])
            }
        }

        stage("Just build api_gateway") {
            when {
                not {
                    branch 'main'
                }
            }
            environment {
                IMAGE_NAME = 'api-gateway-any'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./api_gateway/api_gateway.Dockerfile ./api_gateway"
                        )
                    }
                }
            }
        }
        
        stage('Build & Push api_gateway') {
            when {
                branch 'main'
            }
            environment {
                IMAGE_NAME = 'api-gateway-prod'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./api_gateway/api_gateway.Dockerfile ./api_gateway"
                        )
                        customImage.push()
                        customImage.push('latest')
                    }
                }
            }
        }

        stage("Just build auth_service") {
            when {
                not {
                    branch 'main'
                }
            }
            environment {
                IMAGE_NAME = 'auth-service-any'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./auth_service/auth_service.Dockerfile ./auth_service"
                        )
                    }
                }
            }
        }
        
        stage('Build & Push auth_service') {
            when {
                branch 'main'
            }
            environment {
                IMAGE_NAME = 'auth-service-prod'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./auth_service/auth_service.Dockerfile ./auth_service"
                        )
                        customImage.push()
                        customImage.push('latest')
                    }
                }
            }
        }

        stage("Just build auth_service_migrator") {
            when {
                not {
                    branch 'main'
                }
            }
            environment {
                IMAGE_NAME = 'auth-service-migrator-any'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./auth_service/auth_service_migrator.Dockerfile ./auth_service"
                        )
                    }
                }
            }
        }
        
        stage('Build & Push auth_service_migrator') {
            when {
                branch 'main'
            }
            environment {
                IMAGE_NAME = 'auth-service-migrator-prod'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./auth_service/auth_service_migrator.Dockerfile ./auth_service"
                        )
                        customImage.push()
                        customImage.push('latest')
                    }
                }
            }
        }

        stage("Just build template_service_migrator") {
            when {
                not {
                    branch 'main'
                }
            }
            environment {
                IMAGE_NAME = 'template-service-migrator-any'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./template_service/template_service_migrator.Dockerfile ./template_service"
                        )
                    }
                }
            }
        }
        
        stage('Build & Push template_service_migrator') {
            when {
                branch 'main'
            }
            environment {
                IMAGE_NAME = 'template-service-migrator-prod'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./template_service/template_service_migrator.Dockerfile ./template_service"
                        )
                        customImage.push()
                        customImage.push('latest')
                    }
                }
            }
        }

        stage("Just build template_service") {
            when {
                not {
                    branch 'main'
                }
            }
            environment {
                IMAGE_NAME = 'template-service-any'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./template_service/template_service.Dockerfile ./template_service"
                        )
                    }
                }
            }
        }
        
        stage('Build & Push template_service') {
            when {
                branch 'main'
            }
            environment {
                IMAGE_NAME = 'template-service-prod'
            }
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY_URL}", env.REGISTRY_CREDENTIALS_ID) {
                        def customImage = docker.build("${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}",
                            "-f ./template_service/template_service.Dockerfile ./template_service"
                        )
                        customImage.push()
                        customImage.push('latest')
                    }
                }
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}