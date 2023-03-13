// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
	"reflect"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodecommit"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/pipelines"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdklabs/cdk-nag-go/cdknag/v2"
)

type HelloGoStackProps struct {
	awscdk.StackProps
}

func NewHelloGoPipelineStack(scope constructs.Construct, id string, props *HelloGoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	githubRepoNameStr := ""
	githubRepoName := stack.Node().TryGetContext(jsii.String(id + ":githubRepoName"))
	if githubRepoName != nil {
		githubRepoNameStr = reflect.ValueOf(githubRepoName).String()
	}
	githubConnectionArnStr := ""
	githubConnectionArn := stack.Node().TryGetContext(jsii.String(id + ":githubConnectionArn"))
	if githubConnectionArn != nil {
		githubConnectionArnStr = reflect.ValueOf(githubConnectionArn).String()
	}

	var source pipelines.CodePipelineSource

	if len(githubRepoNameStr) > 0 && len(githubConnectionArnStr) > 0 {
		// https://docs.aws.amazon.com/dtconsole/latest/userguide/connections-create-github.html
		// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/pipelines#readme-codepipeline-sources
		source = pipelines.CodePipelineSource_Connection(jsii.String(githubRepoNameStr), jsii.String("main"), &pipelines.ConnectionSourceOptions{
			ConnectionArn: jsii.String(githubConnectionArnStr),
		})
	} else {
		// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2/awscodecommit#Repository
		repo := awscodecommit.NewRepository(stack, jsii.String("Repository"), &awscodecommit.RepositoryProps{
			RepositoryName: jsii.String("hello-go-cdk"),
			Description:    jsii.String("Hello Go application and CDK stack"),
		})

		awscdk.NewCfnOutput(stack, jsii.String("RepositoryURL"), &awscdk.CfnOutputProps{
			Value:       repo.RepositoryCloneUrlGrc(),
			Description: jsii.String("Repository URL"),
		})

		// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/pipelines#section-readme
		source = pipelines.CodePipelineSource_CodeCommit(repo, jsii.String("main"), &pipelines.CodeCommitSourceOptions{})
	}

	buildCommands := jsii.Strings(
		"npm install -g aws-cdk",
		// "goenv install --list",
		"goenv install 1.19.5",
		"goenv local 1.19.5",
		"go build",
		"cdk synth",
	)

	pipeline := pipelines.NewCodePipeline(stack, jsii.String("CDKPipeline"), &pipelines.CodePipelineProps{
		PipelineName:                 jsii.String("HelloAppCDKPipeline"),
		SelfMutation:                 jsii.Bool(true),
		DockerEnabledForSynth:        jsii.Bool(true),
		DockerEnabledForSelfMutation: jsii.Bool(true),
		Synth: pipelines.NewShellStep(jsii.String("Synth"), &pipelines.ShellStepProps{
			Input:    source,
			Commands: buildCommands,
		}),
	})

	pipeline.AddStage(NewHelloGoAppStage(stack, "HelloGoAppStage", &HelloGoStageProps{}), &pipelines.AddStageOpts{})

	// see https://github.com/cdklabs/cdk-nag/#suppressing-aws-cdk-libpipelines-violations
	pipeline.BuildPipeline()

	AddPipelineStackSuppressions(stack)

	return stack
}

func AddPipelineStackSuppressions(stack awscdk.Stack) {
	suppressions := &[]*cdknag.NagPackSuppression{
		{
			Id:     jsii.String("AwsSolutions-S1"),
			Reason: jsii.String("CDK Pipeline does not need server access logs for Artifacts bucket"),
		},
		{
			Id:     jsii.String("AwsSolutions-CB3"),
			Reason: jsii.String("CDK Pipeline builds docker images"),
		},
		{
			Id:     jsii.String("AwsSolutions-CB4"),
			Reason: jsii.String("CDK Pipeline does not need an AWS KMS key on Artifacts bucket"),
		},
		{
			Id:     jsii.String("AwsSolutions-IAM5"),
			Reason: jsii.String("CDK Pipeline needs full access to Artifacts bucket"),
			AppliesTo: &[]interface{}{
				"Action::s3:List*",
				"Action::s3:GetObject*",
				"Resource::<CDKPipelineArtifactsBucket88CFD064.Arn>/*",
				"Action::s3:GetBucket*",
				"Resource::<CDKPipelineArtifactsBucket88CFD064.Arn>",
			},
		},
		{
			Id:     jsii.String("AwsSolutions-IAM5"),
			Reason: jsii.String("CDK Pipeline needs more wildcard access"),
			AppliesTo: &[]interface{}{
				"Action::s3:*",
				"Action::s3:DeleteObject*",
				"Action::s3:Abort*",
				"Resource::*",
				"Resource::arn:*:iam::<AWS::AccountId>:role/*",
				"Resource::arn:<AWS::Partition>:codebuild:<AWS::Region>:<AWS::AccountId>:report-group/*",
				"Resource::arn:<AWS::Partition>:codebuild:<AWS::Region>:<AWS::AccountId>:report-group/<CDKPipelineBuildSynthCdkBuildProjectE1276695>-*",
				"Resource::arn:<AWS::Partition>:codebuild:<AWS::Region>:<AWS::AccountId>:report-group/<CDKPipelineUpdatePipelineSelfMutation7AA6B177>-*",
				"Resource::arn:<AWS::Partition>:logs:<AWS::Region>:<AWS::AccountId>:log-group:/aws/codebuild/*",
				"Resource::arn:<AWS::Partition>:logs:<AWS::Region>:<AWS::AccountId>:log-group:/aws/codebuild/<CDKPipelineBuildSynthCdkBuildProjectE1276695>:*",
				"Resource::arn:<AWS::Partition>:logs:<AWS::Region>:<AWS::AccountId>:log-group:/aws/codebuild/<CDKPipelineUpdatePipelineSelfMutation7AA6B177>:*",
			},
		},
	}

	cdknag.NagSuppressions_AddStackSuppressions(stack, suppressions, jsii.Bool(true))
}

type HelloGoStageProps struct {
	awscdk.StageProps
}

func NewHelloGoAppStage(scope constructs.Construct, id string, props *HelloGoStageProps) awscdk.Stage {
	var sprops awscdk.StageProps
	if props != nil {
		sprops = props.StageProps
	}
	stage := awscdk.NewStage(scope, &id, &sprops)
	awscdk.Aspects_Of(stage).Add(cdknag.NewAwsSolutionsChecks(&cdknag.NagPackProps{Verbose: jsii.Bool(true)}))

	appStack := NewHelloGoAppStack(stage, "HelloGoAppStack", &HelloGoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})
	AddAppStackSuppressions(appStack)

	return stage
}

func AddAppStackSuppressions(stack awscdk.Stack) {
	suppressions := &[]*cdknag.NagPackSuppression{
		{
			Id:     jsii.String("AwsSolutions-ELB2"),
			Reason: jsii.String("Demo doesn't use access logs"),
		},
		{
			Id:     jsii.String("AwsSolutions-VPC7"),
			Reason: jsii.String("Demo doesn't use VPC flow logs"),
		},
		{
			Id:     jsii.String("AwsSolutions-ECS4"),
			Reason: jsii.String("Demo doesn't use CloudWatch Container Insights"),
		},
		{
			Id:     jsii.String("AwsSolutions-EC23"),
			Reason: jsii.String("Security Group used by ALB with listeners on 80 and 443"),
		},
		{
			Id:     jsii.String("AwsSolutions-IAM5"),
			Reason: jsii.String("TaskDef/ExecutionRole needs wildcard access"),
		},
	}

	cdknag.NagSuppressions_AddStackSuppressions(stack, suppressions, jsii.Bool(true))
}

func NewHelloGoAppStack(scope constructs.Construct, id string, props *HelloGoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	useHttps := stack.Node().TryGetContext(jsii.String(id + ":useHttps"))
	domainName := stack.Node().TryGetContext(jsii.String(id + ":domainName"))
	hostedZoneId := stack.Node().TryGetContext(jsii.String(id + ":hostedZoneId"))
	certificateArn := stack.Node().TryGetContext(jsii.String(id + ":certificateArn"))

	// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/awsecs#readme-task-definitions
	// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2/awsecs#FargateTaskDefinitionProps
	fargateTaskDefinition := awsecs.NewFargateTaskDefinition(stack, jsii.String("TaskDef"), &awsecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(512),
		Cpu:            jsii.Number(256), // 256 (.25 vCPU)
		RuntimePlatform: &awsecs.RuntimePlatform{
			CpuArchitecture: awsecs.CpuArchitecture_X86_64(),
		},
	})

	containerDefinitionOptions := &awsecs.ContainerDefinitionOptions{
		ContainerName: jsii.String("HelloAppContainer"),
		Image: awsecs.ContainerImage_FromAsset(jsii.String("HelloApp"), &awsecs.AssetImageProps{
			Platform: awsecrassets.Platform_LINUX_AMD64(),
		}),
		User: jsii.String("1001"),
	}

	containerDefinitionOptions.HealthCheck = &awsecs.HealthCheck{
		Command: jsii.Strings("CMD-SHELL", "curl -f http://localhost:8080/healthcheck || exit 1"),
	}
	containerDefinitionOptions.Logging = awsecs.LogDriver_AwsLogs(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("HelloApp"),
	})
	container := fargateTaskDefinition.AddContainer(jsii.String("Container"), containerDefinitionOptions)
	container.AddPortMappings(&awsecs.PortMapping{ContainerPort: jsii.Number(8080)})

	applicationLoadBalancedFargateServiceProps := &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		ServiceName:      jsii.String("HelloService"),
		DesiredCount:     jsii.Number(2),
		TaskDefinition:   fargateTaskDefinition,
		Protocol:         awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		TargetProtocol:   awselasticloadbalancingv2.ApplicationProtocol_HTTP,
		LoadBalancerName: jsii.String("hello-alb"),
	}

	if useHttps != nil && reflect.ValueOf(useHttps).Bool() {
		applicationLoadBalancedFargateServiceProps.RedirectHTTP = jsii.Bool(true)
		applicationLoadBalancedFargateServiceProps.Protocol = awselasticloadbalancingv2.ApplicationProtocol_HTTPS
	}

	if domainName != nil {
		var domainNameStr = reflect.ValueOf(domainName).String()
		if domainNameStr != "" {
			applicationLoadBalancedFargateServiceProps.DomainName = jsii.String(domainNameStr)

			if hostedZoneId != nil {
				var hostedZoneIdStr = reflect.ValueOf(hostedZoneId).String()
				if hostedZoneIdStr != "" {
					applicationLoadBalancedFargateServiceProps.DomainZone = awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneAttributes{
						HostedZoneId: jsii.String(hostedZoneIdStr),
						ZoneName:     jsii.String(domainNameStr),
					})
				}
			}
		}
	}

	// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns#readme-application-load-balanced-services
	loadBalancedFargateService := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("FargateService"), applicationLoadBalancedFargateServiceProps)

	if certificateArn != nil {
		var certificateArnStr = reflect.ValueOf(certificateArn).String()
		if certificateArnStr != "" {
			listenerCertificate := awselasticloadbalancingv2.ListenerCertificate_FromArn(jsii.String(certificateArnStr))
			listenerCertificates := []awselasticloadbalancingv2.IListenerCertificate{listenerCertificate}
			loadBalancedFargateService.Listener().AddCertificates(jsii.String("Certificates"), &listenerCertificates)
		}
	}

	albHealthCheck := &awselasticloadbalancingv2.HealthCheck{
		Path:     jsii.String("/healthcheck"),
		Port:     jsii.String("8080"),
		Protocol: awselasticloadbalancingv2.Protocol_HTTP,
	}
	loadBalancedFargateService.TargetGroup().ConfigureHealthCheck(albHealthCheck)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)
	// https://pkg.go.dev/github.com/cdklabs/cdk-nag-go/cdknag/v2
	awscdk.Aspects_Of(app).Add(cdknag.NewAwsSolutionsChecks(&cdknag.NagPackProps{Verbose: jsii.Bool(true)}))

	NewHelloGoPipelineStack(app, "HelloGoPipelineStack", &HelloGoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
