package main

import (
	"cophymaru"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/FePhyFoFum/gophy"
)

func printModel(modelparams []float64, basefreqs []float64) {
	fmt.Fprintln(os.Stderr, "basefreqs -- A:", basefreqs[0], " C:", basefreqs[1], " G:", basefreqs[2], " T:", basefreqs[3])
	fmt.Fprintln(os.Stderr, "modelparams --")
	fmt.Fprintln(os.Stderr, " - ", modelparams[0], modelparams[1], modelparams[2])
	fmt.Fprintln(os.Stderr, modelparams[0], " - ", modelparams[3], modelparams[4])
	fmt.Fprintln(os.Stderr, modelparams[1], modelparams[3], " - ", 1.0)
	fmt.Fprintln(os.Stderr, modelparams[2], modelparams[4], 1.0, "-")
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	//lg := bufio.NewWriter(f)
	tfn := flag.String("t", "", "tree filename")
	afn := flag.String("s", "", "seq filename")
	//anc := flag.Bool("a", false, "calc anc states")
	//stt := flag.Bool("i", false, "calc stochastic time (states will also be calculated)")
	//stn := flag.Bool("n", false, "calc stochastic number (states will also be calculated)")
	//md := flag.Bool("m", false, "model params free")
	//ebf := flag.Bool("b", true, "use empirical base freqs (alt is estimate)")
	wks := flag.Int("w", 4, "number of threads")
	flag.Parse()
	if len(*tfn) == 0 {
		fmt.Fprintln(os.Stderr, "need a tree filename (-t)")
		os.Exit(1)
	}
	if len(*afn) == 0 {
		fmt.Fprintln(os.Stderr, "need a seq filename (-s)")
		os.Exit(1)
	}

	//read a tree file
	//trees := gophy.ReadTreesFromFile(*tfn)
	nwk := cophymaru.ReadLine(*tfn)[0]
	//rt := cophymaru.ReadTree(nwk)
	rt := gophy.ReadNewickString(nwk)
	t := gophy.NewTree()
	t.Instantiate(rt)
	//read a seq file
	nsites := 0
	seqs := map[string][]string{}
	mseqs, numstates := gophy.ReadMSeqsFromFile(*afn)
	seqnames := make([]string, 0)
	for _, i := range mseqs {
		seqs[i.NM] = i.SQs
		seqnames = append(seqnames, i.NM)
		nsites = len(i.SQ)
	}
    x := gophy.NewMultStateModel()
	x.NumStates = numstates
	x.SetMap()
	bf := gophy.GetEmpiricalBaseFreqsMS(mseqs, x.NumStates)
	x.SetBaseFreqs(bf)
	x.EBF = x.BF
    x.SetEqualBF()
    //patterns, patternsint, gapsites, constant, uninformative, _ := gophy.GetSitePatternsMS(mseqs, x)
	_, patternsint, _, _, _, _ := gophy.GetSitePatternsMS(mseqs, x)
	patternval, _ := gophy.PreparePatternVecsMS(t, patternsint, seqs, x)
    //sv := gophy.NewSortedIdxSlice(patternvec)
	//sort.Sort(sv)
	x.SetupQJC()
	fmt.Println(x.Q, x.BF)
	l := gophy.PCalcLogLikePatternsMS(t, x, patternval, *wks)
    fmt.Println(x.NumStates)
	gophy.PCalcSankParsPatternsMultState(t, x, patternval, 1)
	gophy.EstParsBLMultState(t, x, patternval, nsites)
	for _, n := range t.Post {
		n.Len = math.Max(10e-10, n.Len/(float64(nsites)))
	}
	fmt.Println(t.Rt.Newick(true))
	l = gophy.PCalcLogLikePatternsMS(t, x, patternval, *wks)
	fmt.Println(l)
	gophy.OptimizeBLNRMS(t, x, patternval, *wks)
	fmt.Println(t.Rt.Newick(true))
	l = gophy.PCalcLogLikePatternsMS(t, x, patternval, *wks)
	fmt.Println(l)

}
