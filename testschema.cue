foo: int & >= 0 & <100
bar: [string]: #Baz

#Baz: {
	name: string
	blah: string
	foobie: [...int]
}

// This will be provided by the program, not the user.
env: [_]: string
