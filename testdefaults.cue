env: [_]: string
foo: *1 | _
bar: [n=_]: {
	name: n
	blah: *env.SOMEVAR | _
}
