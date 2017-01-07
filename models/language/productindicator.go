package language

import (
	"github.com/RenatoGeh/gospn/spn"
)

// ProductIndicator is a product node that has two children: one is an internal node, and the other
// is an indicator node. This is basically a switch node. If the indicator is set to 1, then this
// node allows the signal to pass downwards. Else it stops here.
type ProductIndicator struct {
	spn.Product
	// Which indicator is attached to this node.
	indicator int
}

// NewProductIndicator creates a new ProductIndicator
func NewProductIndicator(ind int) *ProductIndicator {
	return &ProductIndicator{*spn.NewProduct(), ind}
}

// Soft is a common base for all soft inference methods.
func (p *ProductIndicator) Soft(val spn.VarSet, key string) float64 {
	if _lv, ok := p.Stored(key); ok {
		return _lv
	}

	ch := p.Ch()[0]
	var v float64
	if _y, ok := val[0]; _y == p.indicator || !ok {
		v = ch.Soft(val, key)
	} else {
		ch.Soft(val, key)
		v = 0
	}

	p.Store(key, v)
	return v
}

// Value returns the value of this SPN given a set of valuations.
func (p *ProductIndicator) Value(val spn.VarSet) float64 {
	return p.Soft(val, "soft")
}
