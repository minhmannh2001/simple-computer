package circuit

type NANDGate struct{ output Wire }

func NewNANDGate() *NANDGate { return &NANDGate{output: *NewWire("O", false)} }

func (g *NANDGate) Update(a, b bool) { g.output.Update(!(a && b)) }
func (g *NANDGate) Output() bool     { return g.output.Get() }

type ANDGate struct{ output Wire }

func NewANDGate() *ANDGate { return &ANDGate{output: *NewWire("O", false)} }

func (g *ANDGate) Update(a, b bool) { g.output.Update(a && b) }
func (g *ANDGate) Output() bool     { return g.output.Get() }

type NOTGate struct{ output Wire }

func NewNOTGate() *NOTGate { return &NOTGate{output: *NewWire("O", false)} }

func (g *NOTGate) Update(a bool) { g.output.Update(!a) }
func (g *NOTGate) Output() bool  { return g.output.Get() }

type ORGate struct{ output Wire }

func NewORGate() *ORGate { return &ORGate{output: *NewWire("O", false)} }

// OR via De Morgan: !(NOT a AND NOT b)
func (g *ORGate) Update(a, b bool) { g.output.Update(!(!a && !b)) }
func (g *ORGate) Output() bool     { return g.output.Get() }

type XORGate struct{ output Wire }

func NewXORGate() *XORGate { return &XORGate{output: *NewWire("O", false)} }

// XOR: true only when inputs differ — implemented in NAND-derivable form
func (g *XORGate) Update(a, b bool) { g.output.Update(!((!a && !b) || (a && b))) }
func (g *XORGate) Output() bool     { return g.output.Get() }

type NORGate struct{ output Wire }

func NewNORGate() *NORGate { return &NORGate{output: *NewWire("O", false)} }

func (g *NORGate) Update(a, b bool) { g.output.Update(!a && !b) }
func (g *NORGate) Output() bool     { return g.output.Get() }
