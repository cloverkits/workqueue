package workqueue

type heap struct {
	data []*WaitingFor
}

func (c *heap) Reset() {
	c.data = c.data[:0]
}

func (c *heap) Less(i, j int) bool { return c.data[i].expireAt < c.data[j].expireAt }

func (c *heap) UpdateTTL(ele *WaitingFor, exp int64) {
	var down = exp > ele.expireAt
	ele.expireAt = exp
	if down {
		c.Down(ele.index, c.Len())
	} else {
		c.Up(ele.index)
	}
}

func (c *heap) min(i, j int) int {
	if c.data[i].expireAt < c.data[j].expireAt {
		return i
	}
	return j
}

func (c *heap) Len() int {
	return len(c.data)
}

func (c *heap) Swap(i, j int) {
	c.data[i].index, c.data[j].index = c.data[j].index, c.data[i].index
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

func (c *heap) Push(ele *WaitingFor) {
	ele.index = c.Len()
	c.data = append(c.data, ele)
	c.Up(c.Len() - 1)
}

func (c *heap) Up(i int) {
	var j = (i - 1) >> 2
	if i >= 1 && c.Less(i, j) {
		c.Swap(i, j)
		c.Up(j)
	}
}

func (c *heap) Pop() (ele *WaitingFor) {
	var n = c.Len()
	switch n {
	case 0:
	case 1:
		ele = c.data[0]
		c.data = c.data[:0]
	default:
		ele = c.data[0]
		c.Swap(0, n-1)
		c.data = c.data[:n-1]
		c.Down(0, n-1)
	}
	return
}

func (c *heap) Delete(i int) {
	var n = c.Len()
	switch n {
	case 1:
		c.data = c.data[:0]
	default:
		var down = c.Less(i, n-1)
		c.Swap(i, n-1)
		c.data = c.data[:n-1]
		if i < n-1 {
			if down {
				c.Down(i, n-1)
			} else {
				c.Up(i)
			}
		}
	}
}

func (c *heap) Down(i, n int) {
	var index1 = i<<2 + 1
	if index1 >= n {
		return
	}

	var index2 = i<<2 + 2
	var index3 = i<<2 + 3
	var index4 = i<<2 + 4
	var j = -1

	if index4 < n {
		j = c.min(c.min(index1, index2), c.min(index3, index4))
	} else if index3 < n {
		j = c.min(c.min(index1, index2), index3)
	} else if index2 < n {
		j = c.min(index1, index2)
	} else {
		j = index1
	}

	if j >= 0 && c.Less(j, i) {
		c.Swap(i, j)
		c.Down(j, n)
	}
}

// Front 访问堆顶元素
// Accessing the top Element of the heap
func (c *heap) Front() *WaitingFor {
	return c.data[0]
}