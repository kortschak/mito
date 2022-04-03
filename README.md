# Mito

Mito is a sketch for a message stream processing engine based on [CEL](https://github.com/google/cel-go). Mito provides tools in the lib directory to support collection processing, timestamp handling and other common tasks (see test snippets in [testdata](./testdata) and docs at https://godocs.io/github.com/kortschak/mito/lib).

Some features of mito depend on features that do not yet exist in mainline CEL and some are firmly within the realms of dark arts.

The `mito` command will apply CEL expressions to a JSON value input under the label `data` within the CEL environment. This is intended to be used as a debugging and playground tool.

For example the following CEL expression processes the stream below generating the Cartesian product of the `num` and `let` fields and retaining the original message and adding timestamp metadata.

```
data.map(e, has(e.other) && e.other != '',
	has(e.num) && size(e.num) != 0 && has(e.let) && size(e.let) != 0 ?
		// Handle Cartesian product.
		e.num.map(v1,
			e.let.map(v2,
				e.with({
					"@triggered": now,   // As a value, the start time.
					"@timestamp": now(), // As a function, the time the action happened.
					"original": e.encode_json(),
					"numlet": e.num+e.let,
					"num": v1,
					"let": v2,
				})
		))
	:
		// Handle cases where there is only one of num or let and so
		// the Cartesian product would be empty: S × Ø, S = num or let.
		//
		// This expression is nested to agree with the Cartesian
		// product (an alternative is to flatten that for each e).
		[[e.with({
			"@triggered": now,   // As a value, the start time.
			"@timestamp": now(), // As a function, the time the action happened.
			"original": e.encode_json(),
		})]] 
).flatten().drop_empty().as(res,
	{
		"results": res,
		"timestamps": res.collate('@timestamp').as(t, {
			"first": t.min(),
			"last": t.max(),
			"list": t,
		}),
	}
)
```
working on
```json
[
	{
		"let": ["a", "b"],
		"num": ["1", "2"],
		"other": "random information for first"
	},
	{
		"let": ["aa", "bb"],
		"num": ["12", "22", "33"],
		"other": "random information for second"
	},
	{
		"let": ["a", "b"],
		"num": [],
		"other": "random information for third"
	},
	{
		"let": [], 
		"num": ["1", "2"],
		"other": "random information for fourth"
	},
	{
		"num": ["1", "2"],
		"other": "random information for fifth"
	},
	{
		"let": ["y", "z"],
		"num": ["-1", "-2", "-3"]
	}
]
```
gives
```json
{
	"results": [
		{
			"@timestamp": "2022-04-01T11:56:21.241226877Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "a",
			"num": "1",
			"numlet": [
				"1",
				"2",
				"a",
				"b"
			],
			"original": "{\"let\":[\"a\",\"b\"],\"num\":[\"1\",\"2\"],\"other\":\"random information for first\"}",
			"other": "random information for first"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241265332Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "b",
			"num": "1",
			"numlet": [
				"1",
				"2",
				"a",
				"b"
			],
			"original": "{\"let\":[\"a\",\"b\"],\"num\":[\"1\",\"2\"],\"other\":\"random information for first\"}",
			"other": "random information for first"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241280121Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "a",
			"num": "2",
			"numlet": [
				"1",
				"2",
				"a",
				"b"
			],
			"original": "{\"let\":[\"a\",\"b\"],\"num\":[\"1\",\"2\"],\"other\":\"random information for first\"}",
			"other": "random information for first"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241293719Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "b",
			"num": "2",
			"numlet": [
				"1",
				"2",
				"a",
				"b"
			],
			"original": "{\"let\":[\"a\",\"b\"],\"num\":[\"1\",\"2\"],\"other\":\"random information for first\"}",
			"other": "random information for first"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241312311Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "aa",
			"num": "12",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241618859Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "bb",
			"num": "12",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241677716Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "aa",
			"num": "22",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241703813Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "bb",
			"num": "22",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241740762Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "aa",
			"num": "33",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241762253Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": "bb",
			"num": "33",
			"numlet": [
				"12",
				"22",
				"33",
				"aa",
				"bb"
			],
			"original": "{\"let\":[\"aa\",\"bb\"],\"num\":[\"12\",\"22\",\"33\"],\"other\":\"random information for second\"}",
			"other": "random information for second"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241803009Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"let": [
				"a",
				"b"
			],
			"original": "{\"let\":[\"a\",\"b\"],\"num\":[],\"other\":\"random information for third\"}",
			"other": "random information for third"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241834789Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"num": [
				"1",
				"2"
			],
			"original": "{\"let\":[],\"num\":[\"1\",\"2\"],\"other\":\"random information for fourth\"}",
			"other": "random information for fourth"
		},
		{
			"@timestamp": "2022-04-01T11:56:21.241864081Z",
			"@triggered": "2022-04-01T11:56:21.241224554Z",
			"num": [
				"1",
				"2"
			],
			"original": "{\"num\":[\"1\",\"2\"],\"other\":\"random information for fifth\"}",
			"other": "random information for fifth"
		}
	],
	"timestamps": {
		"first": "2022-04-01T11:56:21.241226877Z",
		"last": "2022-04-01T11:56:21.241864081Z",
		"list": [
			"2022-04-01T11:56:21.241226877Z",
			"2022-04-01T11:56:21.241265332Z",
			"2022-04-01T11:56:21.241280121Z",
			"2022-04-01T11:56:21.241293719Z",
			"2022-04-01T11:56:21.241312311Z",
			"2022-04-01T11:56:21.241618859Z",
			"2022-04-01T11:56:21.241677716Z",
			"2022-04-01T11:56:21.241703813Z",
			"2022-04-01T11:56:21.241740762Z",
			"2022-04-01T11:56:21.241762253Z",
			"2022-04-01T11:56:21.241803009Z",
			"2022-04-01T11:56:21.241834789Z",
			"2022-04-01T11:56:21.241864081Z"
		]
	}
}
```

(Run `mito -data example.json example.cel` to see this locally.)