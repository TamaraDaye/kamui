main_package_path = ./
binary_name = kamui
build_dir = ./bin

.PHONY: help 
help: 
	@echo "Usage:"
	@sed  -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


.PHONY: build
build: 
	@mkdir -p ${build_dir}
	GOARCH=amd64 GOOS=darwin go build -o ${build_dir}/${binary_name}-darwin ${main_package_path}
	GOARCH=amd64 GOOS=linux go build -o ${build_dir}/${binary_name}-linux ${main_package_path}
	GOARCH=amd64 GOOS=darwin go build -o ${build_dir}/${binary_name}-windows.exe ${main_package_path}
	

