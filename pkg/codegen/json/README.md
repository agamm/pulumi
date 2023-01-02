# PCL to JSON (experimental)

This directory contains a PCL to JSON code generator which converts a PCL program into a JSON format that represents the Abstract Syntax Tree (AST) of the program. This JSON can be used by external tools to perform analysis or conversion of PCL programs to target languages without needing to build a specialized PCL parser. The JSON format is not intended to be human readable, but rather to be easily parsed by external tools.

## Usage
Execute `pulumi convert` against a Pulumi YAML program in experimental mode:
```
PULUMI_EXPERIMENTAL=1 pulumi convert --language json --out json
```

Note that Pulumi already type checks the program before converting it to JSON. If the program contains errors, the conversion will fail.

## Format

A PCL program is converted into a JSON object with a field called `nodes` which is an array of the nodes of the program:
```json
{
    "nodes": [...]
}
```
Each `node` can be one of the following:

`Resource` of the following shape:
```json
{
    "type": "Resource",
    "name": "string",
    "token": "string",
    "logicalName": "string",
    "inputs": {
        "string": <expression>
    },
    "options": {
        "string": <expression>
    }
}
```
`LocalVariable` of the following shape
```json
{
    "type": "LocalVariable",
    "name": "string",
    "logicalName": "string",
    "value": <expression>
}
```
`OutputVariable` of the following shape
```json
{
    "type": "OutputVariable",
    "name": "string",
    "logicalName": "string",
    "value": <expression>
}
```
`ConfigVariable` of the following shape
```json
{
    "type": "ConfigVariable",
    "name": "string",
    "logicalName": "string",
    "configType": "<string | number | int | boolean | unknown>",
    "defaultValue": <expression>
}
```
Where each `<expression>` is a JSON object that represents an expression. The shape of the expression depends on the type of expression. For example:
`LiteralValueExpression:`
```json
{
    "type": "LiteralValueExpression",
    "value": "string | number | boolean | null"
}
```
`ObjectConstExpression`: 
```json
{
    "type": "ObjectConstExpression",
    "properties": {
        "string": <expression>
    }
}
```
`TupleConsExpression`: 
```json
{
    "type": "TupleConsExpression",
    "items": [<expression>, ...]
}
```
`FunctionCallExpression`: 
```json
{
    "type": "FunctionCallExpression",
    "name": "string",
    "args": [<expression>, ...]
}
```
This is not the complete list. Consult the code, specifically the `transformExpression` function for the full list of expressions and what they translate to.