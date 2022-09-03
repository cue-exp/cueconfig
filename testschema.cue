#Config: {
	foo: int & >= 0 & <100
	bar: [string]: #Baz
}

#Baz: {
	name: string
	blah: string
	foobie: [...int]
}

#Defaults: {
	config: {
		foo: *1 | _
		bar: [n=_]: {
			name: n
			blah: *"default value" | _
		}
	}
}
