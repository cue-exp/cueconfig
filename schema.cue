#Config: {
	foo: int & >= 0 & <100
	bar: [string]: #Baz
}

#Baz: {
	name: string
	blah: string
	foobie: [...int]
}

config: #Config & {
	foo: *1 | _
	bar: [n=_]: {
		name: n
		blah: *"default value" | _
	}
}
