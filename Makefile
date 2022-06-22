genapitypes:
	curl https://raw.githubusercontent.com/VitalFrog/api_spec/main/openapi.yml > openapi.yaml
	oapi-codegen -old-config-style  -package vfrogapi -generate types ./openapi.yaml > vfrogapi/types.go
	rm openapi.yaml
