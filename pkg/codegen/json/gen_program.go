// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/pulumi/pulumi/pkg/v3/codegen/hcl2/model"
	"github.com/pulumi/pulumi/pkg/v3/codegen/pcl"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/zclconf/go-cty/cty"
)

func transformExpression(expr model.Expression) map[string]interface{} {
	switch expr.(type) {
	case *model.LiteralValueExpression:
		literalExpr := expr.(*model.LiteralValueExpression)
		var value interface{}
		switch literalExpr.Value.Type() {
		case cty.Bool:
			value = literalExpr.Value.True()
		case cty.Number:
			number, _ := literalExpr.Value.AsBigFloat().Float64()
			value = number
		case cty.String:
			value = literalExpr.Value.AsString()
		default:
			value = nil
		}

		return map[string]interface{}{
			"type":  "LiteralValueExpression",
			"value": value,
		}
	case *model.TemplateExpression:
		templateExpression := expr.(*model.TemplateExpression)
		parts := make([]interface{}, len(templateExpression.Parts))
		for i, part := range templateExpression.Parts {
			parts[i] = transformExpression(part)
		}
		return map[string]interface{}{
			"type":  "TemplateExpression",
			"parts": parts,
		}
	case *model.IndexExpression:
		indexExpr := expr.(*model.IndexExpression)
		return map[string]interface{}{
			"type":       "IndexExpression",
			"collection": transformExpression(indexExpr.Collection),
			"key":        transformExpression(indexExpr.Key),
		}
	case *model.ObjectConsExpression:
		objectExpr := expr.(*model.ObjectConsExpression)
		properties := make(map[string]interface{})
		for _, item := range objectExpr.Items {
			if lit, ok := item.Key.(*model.LiteralValueExpression); ok {
				propertyKey := lit.Value.AsString()
				properties[propertyKey] = transformExpression(item.Value)
			}
		}
		return map[string]interface{}{
			"type":       "ObjectConsExpression",
			"properties": properties,
		}
	case *model.TupleConsExpression:
		tupleExpr := expr.(*model.TupleConsExpression)
		items := make([]interface{}, len(tupleExpr.Expressions))
		for i, item := range tupleExpr.Expressions {
			items[i] = transformExpression(item)
		}
		return map[string]interface{}{
			"type":  "TupleConsExpression",
			"items": items,
		}

	case *model.FunctionCallExpression:
		funcExpr := expr.(*model.FunctionCallExpression)
		args := make([]interface{}, len(funcExpr.Args))
		for i, arg := range funcExpr.Args {
			args[i] = transformExpression(arg)
		}
		return map[string]interface{}{
			"type": "FunctionCallExpression",
			"name": funcExpr.Name,
			"args": args,
		}

	case *model.RelativeTraversalExpression:
		traversalExpr := expr.(*model.RelativeTraversalExpression)
		traversal := make([]interface{}, 0)
		for _, part := range traversalExpr.Traversal {
			switch part := part.(type) {
			case hcl.TraverseAttr:
				traversal = append(traversal, map[string]interface{}{
					"type": "TraverseAttr",
					"name": part.Name,
				})
			case hcl.TraverseIndex:
				index, _ := part.Key.AsBigFloat().Int64()
				traversal = append(traversal, map[string]interface{}{
					"type": "TraverseIndex",
					"key":  index,
				})
			}
		}
		return map[string]interface{}{
			"type":      "RelativeTraversalExpression",
			"source":    transformExpression(traversalExpr.Source),
			"traversal": traversal,
		}

	case *model.ScopeTraversalExpression:
		traversalExpr := expr.(*model.ScopeTraversalExpression)
		traversal := make([]interface{}, 0)
		for _, part := range traversalExpr.Traversal {
			switch part := part.(type) {
			case hcl.TraverseAttr:
				traversal = append(traversal, map[string]interface{}{
					"type": "TraverseAttr",
					"name": part.Name,
				})
			case hcl.TraverseIndex:
				index, _ := part.Key.AsBigFloat().Int64()
				traversal = append(traversal, map[string]interface{}{
					"type": "TraverseIndex",
					"key":  index,
				})
			}
		}

		return map[string]interface{}{
			"type":      "ScopeTraversalExpression",
			"rootName":  traversalExpr.RootName,
			"traversal": traversal,
		}

	default:
		return nil
	}
}

func transformResource(resource *pcl.Resource) map[string]interface{} {
	resourceJSON := make(map[string]interface{})
	resourceJSON["type"] = "Resource"
	resourceJSON["name"] = resource.Name()
	resourceJSON["token"] = resource.Token
	resourceJSON["logicalName"] = resource.LogicalName()
	attributes := make(map[string]interface{})
	for _, attr := range resource.Inputs {
		attributes[attr.Name] = transformExpression(attr.Value)
	}
	resourceJSON["attributes"] = attributes
	return resourceJSON
}

func transformLocalVariable(variable *pcl.LocalVariable) map[string]interface{} {
	variableJSON := make(map[string]interface{})
	variableJSON["type"] = "LocalVariable"
	variableJSON["name"] = variable.Name()
	variableJSON["logicalName"] = variable.LogicalName()
	variableJSON["value"] = transformExpression(variable.Definition.Value)
	return variableJSON
}

func transformOutput(output *pcl.OutputVariable) map[string]interface{} {
	outputJSON := make(map[string]interface{})
	outputJSON["type"] = "OutputVariable"
	outputJSON["name"] = output.Name()
	outputJSON["logicalName"] = output.LogicalName()
	outputJSON["value"] = transformExpression(output.Value)
	return outputJSON
}

func transformConfigVariable(variable *pcl.ConfigVariable) map[string]interface{} {
	variableJSON := make(map[string]interface{})
	variableJSON["type"] = "ConfigVariable"
	variableJSON["configType"] = variable.Definition.Type
	variableJSON["name"] = variable.Name()
	variableJSON["logicalName"] = variable.LogicalName()
	return variableJSON
}

func transformProgram(program *pcl.Program) map[string]interface{} {
	programJSON := make(map[string]interface{})
	nodes := make([]interface{}, 0, len(program.Nodes))
	packages := make([]interface{}, 0, len(program.Packages()))
	for _, node := range program.Nodes {
		switch node := node.(type) {
		case *pcl.Resource:
			transformedResource := transformResource(node)
			nodes = append(nodes, transformedResource)
		case *pcl.OutputVariable:
			transformedOutput := transformOutput(node)
			nodes = append(nodes, transformedOutput)
		case *pcl.LocalVariable:
			tranformedVariable := transformLocalVariable(node)
			nodes = append(nodes, tranformedVariable)
		case *pcl.ConfigVariable:
			tranformedVariable := transformConfigVariable(node)
			nodes = append(nodes, tranformedVariable)
		}
	}

	for _, pkg := range program.Packages() {
		packageDef := map[string]interface{}{
			"name":    pkg.Name,
			"version": pkg.Version,
		}

		packages = append(packages, packageDef)
	}

	programJSON["nodes"] = nodes
	programJSON["packages"] = packages
	return programJSON
}

func GenerateProgram(program *pcl.Program) (map[string][]byte, hcl.Diagnostics, error) {
	files := make(map[string][]byte)
	diagnostics := hcl.Diagnostics{}
	programJSON := transformProgram(program)
	programBytes, err := json.MarshalIndent(programJSON, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal program to JSON: %w", err)
	}

	files["program.json"] = programBytes
	return files, diagnostics, nil
}

func GenerateProject(directory string, project workspace.Project, program *pcl.Program) error {
	files, diagnostics, err := GenerateProgram(program)
	if err != nil {
		return err
	}
	if diagnostics.HasErrors() {
		return diagnostics
	}

	for filename, data := range files {
		outPath := path.Join(directory, filename)
		err := ioutil.WriteFile(outPath, data, 0600)
		if err != nil {
			return fmt.Errorf("could not write output program: %w", err)
		}
	}

	return nil
}
