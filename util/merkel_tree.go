package util

import "github.com/cloudflare/cfssl/scan/crypto/sha256"

// 默克尔树

type MerkelTree struct {
	MerkelRootNode *MerkelNode
}

type MerkelNode struct {
	Left  *MerkelNode
	Right *MerkelNode
	Data  []byte
}

// 创建默克尔树
func NewMerkelTree(data [][]byte) *MerkelTree {
	// 首先需要知道data的数量，如果是奇数需要在结尾拷贝一份凑成偶数
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	// 将普通交易计算成默克尔树最远叶节点，保存在切片里
	nodes := []MerkelNode{}
	for i := 0; i < len(data); i++ {
		node := BuildMerkelNode(nil, nil, data[i])
		nodes = append(nodes, node)
	}
	// 循环获得根节点
	for {
		if len(nodes) == 1 {
			break
		}
		newNodes := []MerkelNode{}
		for i := 0; i < len(nodes); i = i + 2 {
			mn := BuildMerkelNode(&nodes[i], &nodes[i+1], nil)
			newNodes = append(newNodes, mn)
		}
		nodes = newNodes
		// 防止奇数叶子结点
		if len(nodes) != 1 && len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}
	}
	return &MerkelTree{
		MerkelRootNode: &nodes[0],
	}
}

// 构建默克尔树节点
func BuildMerkelNode(left, right *MerkelNode, data []byte) MerkelNode {
	if left == nil && right == nil {
		datum := sha256.Sum256(data)
		mn := MerkelNode{nil, nil, datum[:]}
		return mn
	}
	sumData := append(left.Data, right.Data...)
	finalData := sha256.Sum256(sumData)
	node := MerkelNode{left, right, finalData[:]}
	return node
}
