package learn

import (
	"fmt"
	//"math"
	"sort"

	spn "github.com/RenatoGeh/gospn/src/spn"
	utils "github.com/RenatoGeh/gospn/src/utils"
)

// We refer to this structural learning algorithm as the Gens Algorithm for structural learning.
// The full article describing this algorithm schema can be found at:
//
// 		http://spn.cs.washington.edu/pubs.shtml
//
// Under the name of
//
// 		Learning the Structure of Sum-Product Networks
// 		Robert Gens and Pedro Domingos; ICML 2013
//
// For clustering we use k-means clustering. Our implementation can be seen in file:
//
// 		/src/utils/kmeans.go
//
// As for testing the independency between two variables we use the Chi-Square independence test,
// present in file:
//
// 		/src/utils/indtest.go
//
// Function Gens takes as input a matrix of data instances, where the columns are variables and
// lines are the observed instantiations of each variable.
//
// 		+-----+------------------------------+
// 		|     | X_1   X_2   X_3   ...   X_n  |
// 		+-----+------------------------------+
// 		| I_1 | x_11  x_12  x_13  ...   x_1n |
// 		| I_2 | x_21  x_22  x_23  ...   x_2n |
// 		|  .  |  .     .     .     .     .   |
// 		|  .  |  .     .     .     .     .   |
// 		|  .  |  .     .     .     .     .   |
// 		| I_m | x_m1  x_m2  x_m3  ...   x_mn |
// 		+-----+------------------------------+
//
// Where X={X_1,...,X_n} is the set of variables and I={I_1,...,I_m} is the set of instances.
// Each x_ij is the i-th observed instantiation of X_j.
func Gens(sc map[int]Variable, data []map[int]int) spn.SPN {
	n := len(sc)

	fmt.Printf("Sample size: %d, scope size: %d\n", len(data), n)

	// If the data's scope is unary, then we return a leaf (i.e. a univariate distribution).
	if n == 1 {
		fmt.Println("Creating new leaf...")

		// m number of instantiations.
		m := len(data)
		// pr is the univariate probability distribution.
		var tv *Variable
		for _, v := range sc {
			tv = &v
		}
		pr, l := make([]float64, tv.Categories), tv.Categories
		for i := 0; i < m; i++ {
			pr[data[i][tv.Varid]]++
		}
		for i := 0; i < l; i++ {
			pr[i] /= float64(m)
		}

		leaf := spn.NewUnivDist(tv.Varid, pr)
		fmt.Println("Leaf created.")
		return leaf
	}

	// Else we check for independent subsets of variables. We separate variables in k partitions,
	// where every partition is pairwise indepedent with another.
	fmt.Println("Preparing to create new product node...")

	// vdata is the transpose of data.
	fmt.Println("Creating VarDatas for Independency Test...")
	vdata, l := make([]*utils.VarData, n), 0
	for _, v := range sc {
		tn := len(data)
		// tdata is the transpose of data[k].
		tdata := make([]int, tn)
		for j := 0; j < tn; j++ {
			tdata[j] = data[j][v.Varid]
		}
		vdata[l] = utils.NewVarData(v.Varid, v.Categories, tdata)
		l++
	}

	fmt.Println("Creating new Independency graph...")
	// Independency graph.
	igraph := utils.NewIndepGraph(vdata)

	// If true, then we can partition the set of variables in data into independent subsets. This
	// means we can create a product node (since product nodes' children have disjoint scopes).
	if len(igraph.Kset) > 1 {
		fmt.Println("Found independency between variables. Creating new product node...")
		// prod is the new product node. m is the number of disjoint sets. kset is a shortcut.
		prod, m, kset := spn.NewProduct(), len(igraph.Kset), &igraph.Kset
		tn := len(data)
		for i := 0; i < m; i++ {
			// Data slices of the relevant vectors.
			tdata := make([]map[int]int, tn)
			// Number of variables in set of variables kset[i].
			s := len((*kset)[i])
			for j := 0; j < tn; j++ {
				tdata[j] = make(map[int]int)
				for l := 0; l < s; l++ {
					// Get the instanciations of variables in kset[i].
					//fmt.Printf("[%d][%d] => %v vs %v | %v vs %v\n", j, k, (*kset)[i][k], len(data[j]), len(tdata[j]), k)
					k := (*kset)[i][l]
					tdata[j][k] = data[j][k]
				}
			}
			// Create new scope with new variables.
			nsc := make(map[int]Variable)
			for j := 0; j < s; j++ {
				t := (*kset)[i][j]
				nsc[t] = Variable{t, sc[t].Categories}
			}
			fmt.Printf("LENGTH: %d\n", len(tdata))
			fmt.Println("Product node created. Recursing...")
			// Adds the recursive calls as children of this new product node.
			prod.AddChild(Gens(nsc, tdata))
		}
		return prod
	}

	// Else we perform k-clustering on the instances.
	fmt.Println("No independency found. Preparing for clustering...")
	sum := spn.NewSum()

	m := len(data)
	mdata := make([][]int, m)
	for i := 0; i < m; i++ {
		lc := len(data[i])
		mdata[i] = make([]int, lc)
		l := 0
		keys := make([]int, lc)
		for k, _ := range data[i] {
			keys[l] = k
			l++
		}
		sort.Ints(keys)
		l = 0
		for j := 0; j < lc; j++ {
			mdata[i][j] = data[i][keys[j]]
		}
	}

	fmt.Printf("data: %d, mdata: %d\n", len(data), len(mdata))
	const KClusters = 2
	if len(mdata) < KClusters {
		// Fully factorized form.
		// All instances are approximately the same.
		m := len(data)
		for _, v := range sc {
			pr, l := make([]float64, v.Categories), v.Categories
			for i := 0; i < m; i++ {
				pr[data[i][v.Varid]]++
			}
			for i := 0; i < l; i++ {
				pr[i] /= float64(m)
			}
			leaf, w := spn.NewUnivDist(v.Varid, pr), 1.0/float64(n)
			sum.AddChildW(leaf, w)
		}
		return sum
	}
	clusters := utils.KMeansV(KClusters, mdata)
	k := len(clusters)

	emptyc := 0
	var empties []int
	for i := 0; i < k; i++ {
		if len(clusters[i]) == 0 {
			emptyc++
			empties = append(empties, i)
		}
	}

	if emptyc == k-1 {
		// All instances are approximately the same.
		m := len(data)
		for _, v := range sc {
			pr, l := make([]float64, v.Categories), v.Categories
			for i := 0; i < m; i++ {
				pr[data[i][v.Varid]]++
			}
			for i := 0; i < l; i++ {
				pr[i] /= float64(m)
			}
			leaf, w := spn.NewUnivDist(v.Varid, pr), 1.0/float64(n)
			sum.AddChildW(leaf, w)
		}
		return sum
	}

	fmt.Println("Reformating clusters to appropriate format and creating sum nodes...")
	e := 0
	for i := 0; i < k; i++ {
		if len(empties) > 0 && i == empties[e] {
			e++
			continue
		}

		s := len(clusters[i])
		ndata := make([]map[int]int, s)
		for j := 0; j < s; j++ {
			ndata[j] = make(map[int]int)
			// k is indices in original data. v is instance in data[k].
			for k, _ := range clusters[i] {
				for vi, inst := range data[k] {
					ndata[j][vi] = inst
				}
			}
		}
		fmt.Println("Created new sum node. Recursing...")
		sum.AddChildW(Gens(sc, ndata), float64(s)/float64(n))
	}

	return sum
}