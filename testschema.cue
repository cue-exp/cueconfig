#Config: {
	runtime: #Runtime
	foo: int & >= 0 & <100
	bar: [string]: #Baz
}

#Baz: {
	name: string
	blah: string
	foobie: [...int]
}

#Runtime: env: [_]: string

#Defaults: {
	runtime: #Runtime
	config: {
		foo: *1 | _
		bar: [n=_]: {
			name: n
			blah: *runtime.env.SOMEVAR | _
		}
	}
}
