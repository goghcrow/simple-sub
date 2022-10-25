# simple-sub in golang

## example

```go
func TestPgrm(t *testing.T) {
	var typer = NewTyper()

	for _, tt := range []struct {
		name     string
		pgrm     string
		expected []string
	}{
		{
			"mlsub", // from https://www.cl.cam.ac.uk/~sd601/mlsub/
			`
	// 单行注释
	/*
	多行注释
	*/
	let id = fun x -> x
	let twice = fun f x -> f( f( x ) )
	let object1 = { x: 42, y: id }
	let object2 = { x: 17, y: false }
	let pick_an_object = fun b -> if b then object1 else object2
	let rec recursive_monster = fun x -> { thing: x, self: recursive_monster(x) }
	`,
			[]string{
				"'a -> 'a",
				"('a ∨ 'b -> 'b) -> 'a -> 'b",
				"{x: int, y: 'a -> 'a}",
				"{x: int, y: bool}",
				"bool -> {x: int, y: bool ∨ ('a -> 'a)}",
				"'a -> {self: 'b, thing: 'a} as 'b",
			},
		},
		{
			"top-level-polymorphism",
			`
	let id = fun x -> x
	let ab = {u : id( 0 ), v : id( true ) }
`,
			[]string{
				"'a -> 'a",
				"{u: int, v: bool}",
			},
		},
		{
			"rec-producer-consumer",
			`
	  let rec produce = fun arg -> { head : arg, tail : produce( succ( arg ) ) }
      let rec consume = fun strm -> add( strm.head,  consume( strm.tail ) )
      
      let codata = produce(42)
      let res = consume(codata)
      
      let rec codata2 = { head : 0, tail : { head : 1, tail : codata2 } }
      let res = consume( codata2 )
      
      let rec produce3 = fun b -> { head : 123, tail : if b then codata else codata2 }
      let res = fun x -> consume( produce3( x ) )
      
      let consume2 =
        let rec go = fun strm -> add( strm.head, add( strm.tail.head, go( strm.tail.tail ) ) )
        in fun strm -> add( strm.head, go( strm.tail ) )
        // in go
      // let rec consume2 = fun strm -> add( strm.head, add( strm.tail.head, consume2( strm.tail.tail ) ) )
      let res = consume2( codata2 )
`,
			[]string{
				"int -> {head: int, tail: 'a} as 'a",
				"{head: int, tail: 'a} as 'a -> int",
				"{head: int, tail: 'a} as 'a",
				"int",
				"{head: int, tail: {head: int, tail: 'a}} as 'a",
				"int",
				"bool -> {head: int, tail: {head: int, tail: 'a}} as 'a",
				// ^ simplifying this would probably require more advanced
				// automata-based techniques such as the one proposed by Dolan
				"bool -> int",
				"{head: int, tail: {head: int, tail: 'a}} as 'a -> int",
				"int",
			},
		},
		{
			"misc",
			`
      // 
      // From a comment on the blog post:
      // 
      let rec r = fun a -> r
      let join = fun a -> fun b -> if true then a else b
      let s = join ( r, r )
      // 
      // Inspired by [Pottier 98, chap 13.4]
      // 
      let rec f = fun x -> fun y -> add( f( x.tail, y ), f( x,  y ) )
      let rec f = fun x -> fun y -> add( f( x.tail, y ), f( y,  x ) )
      let rec f = fun x -> fun y -> add( f( x.tail, y ), f( x, y.tail ) )
      let rec f = fun x -> fun y -> add( f( x.tail, y.tail ), f( x.tail, y.tail ) )
      let rec f = fun x -> fun y -> add( f( x.tail, x.tail ), f( y.tail, y.tail ) )

	  let rec f = fun x -> fun y -> add( f( x.tail, x ), f( y.tail, y ) )
	  let rec f = fun x -> fun y -> add( f( x.tail, y ), f( y.tail, x ) )
      // 
      let f = fun x -> fun y -> if true then { l : x, r : y } else { l : y, r : x } // 2-crown
      // 
      // Inspired by [Pottier 98, chap 13.5]
      // 
      let rec f = fun x -> fun y -> if true then x else { t : f( x.t, y.t ) }
`,
			[]string{
				"(⊤ -> 'a) as 'a",
				"'a -> 'a -> 'a",
				"(⊤ -> 'a) as 'a",

				"{tail: 'a} as 'a -> ⊤ -> int",
				"{tail: 'a} as 'a -> {tail: 'b} as 'b -> int",
				"{tail: 'a} as 'a -> {tail: 'b} as 'b -> int",
				"{tail: 'a} as 'a -> {tail: 'b} as 'b -> int",
				"{tail: {tail: 'a} as 'a} -> {tail: {tail: 'b} as 'b} -> int",
				// ^ Could simplify more `{tail: {tail: 'a} as 'a}` to `{tail: 'a} as 'a`
				//    This would likely require another hash-consing pass.
				//    Indeed, currently, we coalesce {tail: ‹{tail: ‹α25›}›} and it's hash-consing
				//    which introduces the 'a to stand for {tail: ‹α25›}
				// ^ Note: MLsub says:
				//    let rec f = fun x -> fun y -> (f x.tail x) + (f y.tail y)
				//    val f : ({tail : (rec b = {tail : b})} -> ({tail : {tail : (rec a = {tail : a})}} -> int))
				"{tail: 'a} as 'a -> {tail: {tail: 'b} as 'b} -> int",
				// ^ Note: MLsub says:
				//    let rec f = fun x -> fun y -> (f x.tail x.tail) + (f y.tail y.tail)
				//    val f : ({tail : {tail : (rec b = {tail : b})}} -> ({tail : {tail : (rec a = {tail : a})}} -> int))
				"{tail: 'a} as 'a -> {tail: {tail: 'b} as 'b} -> int",

				"'a -> 'a -> {l: 'a, r: 'a}",

				"('b ∧ {t: 'a}) as 'a -> {t: 'c} as 'c -> ('b ∨ {t: 'd}) as 'd",
				// ^ Note: MLsub says:
				//    let rec f = fun x -> fun y -> if true then x else { t = f x.t y.t }
				//    val f : (({t : (rec d = ({t : d} & a))} & a) -> ({t : (rec c = {t : c})} -> ({t : (rec b = ({t : b} | a))} | a)))
				// ^ Pottier says a simplified version would essentially be, once translated to MLsub types:
				//    {t: 'a} as 'a -> 'a -> {t: 'd} as 'd
				// but even he does not infer that.
				// Notice the loss of connection between the first parameetr and the result, in his proposed type,
				// which he says is not necessary as it is actually implied.
				// He argues that if 'a <: F 'a and F 'b <: 'b then 'a <: 'b, for a type operator F,
				// which does indeed seem true (even in MLsub),
				// though leveraging such facts for simplification would require much more advanced reasoning.
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pgrm := parsePgrm(tt.pgrm)
			tyv, err := typer.inferTypes(pgrm, typer.Builtins())
			if err != nil {
				panic(err)
			}
			for i, poly := range tyv {
				t.Log(pgrm.Defs[i])
				ty := poly.instantiate(typer, 0)
				cty := typer.canonicalizeType(ty)
				sty := typer.simplifyType(cty)
				res := typer.coalesceCompactType(sty).Show()
				t.Log(res)
				t.Log("=====================")
				if res != tt.expected[i] {
					t.Errorf("expect %s actual %s", tt.expected, res)
				}
			}
		})
	}
}
```

## ref

- [The Simple Essence of Algebraic Subtyping](https://lptk.github.io/simple-sub-paper)
- https://github.com/LPTK/simple-sub