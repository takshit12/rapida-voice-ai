GO_PROJECT_MODULE="github.com/rapidaai/protos"
OUT_DIR="/protos/"
rm -rf ./protos/*.go
find ./api/document-api/app/bridges/artifacts/protos -name "*_pb2.py" -delete
find ./api/document-api/app/bridges/artifacts/protos -name "*_pb2_grpc.py" -delete
find ./api/document-api/app/bridges/artifacts/protos -name "*_pb2.pyi" -delete


protoc -I=./protos/artifacts/ --go_opt=module="${GO_PROJECT_MODULE}" --go_out=."${OUT_DIR}" --go-grpc_opt=module="${GO_PROJECT_MODULE}" --go-grpc_out=require_unimplemented_servers=false:."${OUT_DIR}" ./protos/artifacts/*.proto

python3 -m grpc.tools.protoc \
    -I ./protos/artifacts \
    --pyi_out=./api/document-api/app/bridges/artifacts/protos \
    --python_out=./api/document-api/app/bridges/artifacts/protos \
    --grpc_python_out=./api/document-api/app/bridges/artifacts/protos \
    ./protos/artifacts/*.proto


find "api/document-api/app/bridges/artifacts/protos/" \
  -type f \( -name "*.py" -o -name "*.pyi" \) \
  -exec sed -i.bak -E \
    '/^import [a-zA-Z0-9_]+_pb2/ s|import ([a-zA-Z0-9_]+_pb2)|import app.bridges.artifacts.protos.\1|' {} +

# Remove backup files
find "api/document-api/app/bridges/artifacts/protos/" -name "*.bak" -delete
