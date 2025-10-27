module testdata

go 1.25

replace multimodule_implements => ./multimodule_implements

replace multimodule_immutable => ./multimodule_immutable

replace multimodule_constructor => ./multimodule_constructor

replace multimodule_testonly => ./multimodule_testonly

replace multimodule_ignore => ./multimodule_ignore

require (
	multimodule_constructor v0.0.0-00010101000000-000000000000 // indirect
	multimodule_ignore v0.0.0-00010101000000-000000000000 // indirect
	multimodule_immutable v0.0.0-00010101000000-000000000000 // indirect
	multimodule_implements v0.0.0-00010101000000-000000000000 // indirect
	multimodule_testonly v0.0.0-00010101000000-000000000000 // indirect
)
