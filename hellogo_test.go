// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestHelloGoPipelineStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewHelloGoPipelineStack(app, "HelloGoPipelineStack", nil)

	// THEN
	template := assertions.Template_FromStack(stack, &assertions.TemplateParsingOptions{})

	template.HasResourceProperties(jsii.String("AWS::CodePipeline::Pipeline"), map[string]interface{}{
		"Name": "HelloAppCDKPipeline",
	})
}
