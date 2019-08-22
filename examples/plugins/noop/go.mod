module noop

replace github.com/facebookincubator/go2chef => ../../../

go 1.12

require (
	github.com/facebookincubator/go2chef v0.0.0-00010101000000-000000000000
	github.com/mitchellh/mapstructure v1.1.2
	github.com/oko/logif v0.0.0-20190820152755-d4b69729d8ad
)
