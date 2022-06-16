genapitypes:
	curl https://raw.githubusercontent.com/VitalFrog/api_spec/main/openapi.yml > openapi.yaml
	oapi-codegen -old-config-style  -package api_client -generate types ./openapi.yaml > api_client/types.go
	rm openapi.yaml
