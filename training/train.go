package main

import (
	"flag"
	"fmt"
	"lda"
	"rand"
	"time"
)

var (
	num_topics = flag.Int("num_topics", 2, "The number of topics expected in trained model")
	topic_prior = flag.Float64("topic_prior", 0.1, "The parameter of symmetric Dirichlet on topics")
	word_prior = flag.Float64("word_prior", 0.01, "The parameter of symmetric Dirichlet on words")
	corpus_file = flag.String("corpus_file", "", "The (input) training data file")
	model_file = flag.String("model_file", "", "The (output) model file")
        burn_in_iterations = flag.Int("burn_in_iterations", 50,
		"The number of Gibbs sampling iterations for burning in the MCMC")
        accumulate_iterations = flag.Int("accumulate_iterations", 10,
		"The number of Gibbs sampling iterations for accumulating the sampling results")
        compute_loglikelihood = flag.Bool("compute_loglikelihood", true,
		"Whether to compute and output the likelihood after each Gibbs sampling iteration")
)

func CheckFlagsValid() bool {
	valid := true
	if *num_topics < 2 {
		fmt.Println("num_topics must be larger than or equal to 2")
		valid = false
	}
	if *topic_prior <= 0 {
		fmt.Println("topic_prior must be positive")
		valid = false
	}
	if *word_prior <= 0 {
		fmt.Println("word_prior must be positive")
		valid = false
	}
	if len(*corpus_file) == 0 {
		fmt.Println("coprus_file must be specified")
		valid = false
	}
	if len(*model_file) == 0 {
		fmt.Println("model_file must be speicfied")
		valid = false
	}
	if *burn_in_iterations <= 0 {
		fmt.Println("burn_in_iterations must be positive")
		valid = false
	}
	if *accumulate_iterations <= 0 {
		fmt.Println("accumulate_iterations must be positive")
		valid = false
	}
	return valid
}

func main() {
	flag.Parse()
	if !CheckFlagsValid() {
		fmt.Printf("Stop training due to invalid flag setting.\n")
		return
	}

	rand.Seed(time.Nanoseconds())

	corpus, err := lda.LoadCorpus(*corpus_file, 2)
	if err != nil {
		fmt.Printf("Error in loading: " + *corpus_file + ", due to " + err.String())
		return
	}

	model := lda.CreateModel(*num_topics, corpus)
	accum_model := lda.NewModel(*num_topics)
	sampler := lda.NewSampler(*topic_prior, *word_prior, model, accum_model)

	for iter := 0; iter < *burn_in_iterations + *accumulate_iterations; iter++ {
		fmt.Printf("Iteration %d ... ", iter)
		if (*compute_loglikelihood) {
			fmt.Printf("log-likelihood: %f\n", sampler.CorpusLogLikelihood(corpus))
		} else {
			fmt.Printf("\n")
		}
		sampler.CorpusGibbsSampling(corpus, true, iter < *burn_in_iterations)
	}

	if err := accum_model.SaveModel(*model_file); err != nil {
		fmt.Printf("Cannot save model due to " + err.String())
	}

	return
}

