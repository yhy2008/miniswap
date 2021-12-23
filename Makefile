contracts: contracts/Token.sol abigenBindings/abi/Token.abi abigenBindings/bin/Token.bin \
 					 contracts/Exchange.sol abigenBindings/abi/Exchange.abi abigenBindings/bin/Exchange.bin \
 					 contracts/Factory.sol abigenBindings/abi/Factory.abi abigenBindings/bin/Factory.bin
	truffle compile
	truffle run abigen Token Exchange Factory
	abigen --abi=abigenBindings/abi/Token.abi --bin=abigenBindings/bin/Token.bin --pkg=token --out=gobinding/token/token.go
	abigen --abi=abigenBindings/abi/Exchange.abi --bin=abigenBindings/bin/Exchange.bin --pkg=exchange --out=gobinding/exchange/exchange.go
	abigen --abi=abigenBindings/abi/Factory.abi --bin=abigenBindings/bin/Factory.bin --pkg=factory --out=gobinding/factory/factory.go
