// Package gen เก็บ type ที่ generate จาก api/openapi.yaml ด้วย oapi-codegen — ห้ามแก้ไฟล์ *.gen.go ด้วยมือ
// regenerate ด้วย `make generate` (CI ตรวจว่าไฟล์ตรงกับ contract)
package gen

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config ../../../oapi-codegen.yaml ../../../api/openapi.yaml
