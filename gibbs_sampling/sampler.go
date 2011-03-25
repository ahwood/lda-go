package lda

import (
	"fmt"
	"math"
)

type Sampler struct {
	topic_prior float64
	word_prior  float64
	model       *Model
	accum_model *Model
}

func NewSampler(topic_prior float64, word_prior float64, model *Model, accum_model *Model) *Sampler {
	return &Sampler{topic_prior, word_prior, model, accum_model}
}

func (sampler *Sampler) GenerateTopicDistributionForWord(doc *Document,
	word string, target_topic int, update_model bool) Distribution {
	num_topics := sampler.model.NumTopics()
	num_words := sampler.model.NumWords()
	distribution := NewDistribution(num_topics)
	word_histogram := sampler.model.GetWordTopicHistogram(word)

	for k := 0; k < num_topics; k++ {
		// We will need to temporarily unassign the word from its old
		// topic, which we accomplish by decrementing the appropriate
		// counts by 1.
		adjustment := 0
		if update_model && k == target_topic {
			adjustment = -1
		}
		topic_word_factor := float64(word_histogram[k] + adjustment)
		global_topic_factor := float64(sampler.model.GetGlobalTopicHistogram()[k] + adjustment)
		document_topic_factor := float64(doc.topic_histogram[k] + adjustment)
		distribution[k] = (topic_word_factor + sampler.word_prior) *
                        (document_topic_factor + sampler.topic_prior) /
                        (global_topic_factor + float64(num_words) * sampler.word_prior)
	}
	return distribution
}

func (sampler *Sampler) DocumentGibbsSampling(doc *Document, update_model bool) {
	for iter, _ := NewWordIterator(doc); !iter.Done(); iter.Next() {
		// This is a (non-normalized) probability distribution from which we will
		// select the new topic for the current word occurrence.
		new_topic_distribution := sampler.GenerateTopicDistributionForWord(
			doc, iter.Word(), iter.Topic(), update_model)
		new_topic := GetAccumulativeSample(new_topic_distribution);
		if new_topic != -1 {
			// If new_topic != -1 (i.e. GetAccumulativeSample) runs OK, we
			// update document and model parameters with the new topic.
			if (update_model) {
				sampler.model.ReassignTopic(iter.Word(), iter.Topic(), new_topic)
			}
			iter.SetTopic(new_topic);
		} else {
			panic(fmt.Sprintf("Cannot sample from: %v", new_topic_distribution))
		}
	}
}

func (sampler *Sampler) CorpusGibbsSampling(corpus *Corpus, update_model bool, burn_in bool) {
	for _, doc := range *corpus {
		sampler.DocumentGibbsSampling(doc, update_model)
	}

	if sampler.accum_model != nil && update_model && !burn_in {
		sampler.accum_model.AccumulateModel(sampler.model)
	}
}

func (sampler *Sampler) DocumentLogLikelihood(doc *Document) float64 {
	num_topics := sampler.model.NumTopics()
	doc_length := doc.Length()

	// Compute P(z|d) for the given document and all topics.
	prob_topic_given_document := NewDistribution(num_topics)
	smoothed_doc_length := float64(doc_length) + sampler.topic_prior * float64(num_topics)
	for i, v := range doc.topic_histogram {
		prob_topic_given_document[i] = (float64(v) + sampler.topic_prior) / smoothed_doc_length
	}

	// Get global topic occurrences, which will be used to compute P(w|z).
	global_topic_histogram := sampler.model.GetGlobalTopicHistogram()
	prob_word_given_topic := NewDistribution(num_topics)
	log_likelihood := 0.0;

	// A document's log-likelihood is the sum of log-likelihoods
	// of its words.  Compute the likelihood for every word and
	// sum the logs.
	for iter, _ := NewWordIterator(doc); !iter.Done(); iter.Next() {
		// Get topic_count_distribution of the current word,
		// which will be used to Compute P(w|z).
		word_topic_histogram := sampler.model.GetWordTopicHistogram(iter.Word())

		// Compute P(w|z).
		for t := 0; t < num_topics; t++ {
			prob_word_given_topic[t] =
				(float64(word_topic_histogram[t]) + sampler.word_prior) /
				(float64(global_topic_histogram[t]) + float64(doc.Length()) * sampler.word_prior)
		}

		// Compute P(w) = sum_z P(w|z)P(z|d)
		prob_word := 0.0
		for t := 0; t < num_topics; t++ {
			prob_word += prob_word_given_topic[t] * prob_topic_given_document[t]
		}
		log_likelihood += math.Log(prob_word);
	}

	return log_likelihood
}

func (sampler *Sampler) CorpusLogLikelihood(corpus *Corpus) float64 {
	total_log_likelihood := 0.0
	for _, v := range *corpus {
		total_log_likelihood += sampler.DocumentLogLikelihood(v)
	}
	return total_log_likelihood
}
