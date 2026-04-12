module github.com/fipaan/ap2-uni/op-assign

go 1.25.5

replace github.com/fipaan/lib.go/string => ../../lib.go/string

replace github.com/fipaan/lib.go/path => ../../lib.go/path

replace github.com/fipaan/nob.go => ../../nob.go

require github.com/fipaan/nob.go v0.0.0-00010101000000-000000000000

require (
	github.com/fipaan/lib.go/path v0.0.0-00010101000000-000000000000 // indirect
	github.com/fipaan/lib.go/string v0.0.0-00010101000000-000000000000
)
