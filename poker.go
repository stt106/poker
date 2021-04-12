package poker

// this models such a poker game that:
// 1. no joker i.e. no wild card or five kind
// 2. aces can start a straight (smallest straight) or end a straight.
// 3. when two hands tie on its category, then winning hand is decided based on its associated category card rank, for instance
// 44445 beats 33336 because 4 > 3; similarly 22885 beats 44779 because 8 > 7.
// 4. multiple decks are allowed, so using rule 3, 88889 beats 88887.
// 5. the output must be in the original hand input format.

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type category int

const (
	// hand category in descreasing order.
	straightFlush category = iota
	fourKind
	fullHouse
	flush
	straight
	threeKind
	twoPair
	onePair
	highCard
)

type card struct {
	rank int
	suit string
}

type hand struct {
	cat           category
	highestRank   int
	rankSum       int
	originalIndex int // refers to the original input hands index
}

func getHighestRank(cards []card) int {
	hr := cards[0].rank
	for _, c := range cards[1:] {
		if c.rank > hr {
			hr = c.rank
		}
	}
	return hr
}

// BestHand finds the winning hand(s) for the poker game where single or multiple decks are allowed but no jokers.
// It must output the winning hand(s) in the original input format.
func BestHand(hands []string) ([]string, error) {
	chs := make([]hand, 0, len(hands))
	for i, hand := range hands {
		cards, err := validateHand(hand)
		if err != nil {
			return nil, err
		}
		if len(hands) == 1 { // single valid hand wins.
			return hands, nil
		}
		ch := categoriseHand(cards)
		ch.originalIndex = i // store the original hand index for final result output.
		chs = append(chs, ch)
	}
	sort.Slice(chs, func(i, j int) bool {
		if chs[i].cat == chs[j].cat {
			if chs[i].highestRank == chs[j].highestRank {
				// when highest rank also ties, sort based on other card ranks, down to the last card.
				// This equivalents to sorting based on rank sum.
				return chs[i].rankSum > chs[j].rankSum
			}
			// same category hand checks highest rank card.
			return chs[i].highestRank > chs[j].highestRank
		}
		// firstly sorting based on hand category
		return chs[i].cat < chs[j].cat
	})

	indices := []int{chs[0].originalIndex} // first one should be the winner.
	// check for multiple winners.
	for _, hc := range chs[1:] {
		if hc.cat == chs[0].cat && hc.rankSum == chs[0].rankSum && hc.highestRank == chs[0].highestRank {
			indices = append(indices, hc.originalIndex)
		}
	}
	res := []string{}
	for _, i := range indices {
		res = append(res, hands[i])
	}
	return res, nil
}

// valiateHand valiates a given hand and if ok parse it into a card slice.
func validateHand(hand string) ([]card, error) {
	hands := strings.Split(hand, " ")
	if len(hands) != 5 {
		return nil, errors.New("invalid hand length")
	}
	cards := make([]card, 0, 5)
	validateRank := func(rank string) (int, error) {
		switch rank {
		case "A":
			return 14, nil
		case "J":
			return 11, nil
		case "Q":
			return 12, nil
		case "K":
			return 13, nil
		default:
			r, err := strconv.Atoi(rank)
			if r < 2 || r > 10 || err != nil {
				return 0, fmt.Errorf("invalid rank:%w", err)
			}
			return r, nil
		}
	}
loop:
	for _, c := range hands {
		for _, s := range []string{"♢", "♧", "♡", "♤"} {
			if strings.HasSuffix(c, s) {
				c := strings.TrimSuffix(c, s)
				if r, err := validateRank(c); err != nil {
					return nil, errors.New("invalid rank")
				} else {
					cards = append(cards, card{rank: r, suit: s})
					continue loop
				}
			}
		}
		return nil, errors.New("invalid suit")
	}
	return cards, nil
}

func categoriseHand(cards []card) hand {
	sum := cards[0].rank + cards[1].rank + cards[2].rank + cards[3].rank + cards[4].rank
	groups := make(map[int][]card, 5)
	// grouping cards based its rank so that we can count each rank group.
	for _, c := range cards {
		groups[c.rank] = append(groups[c.rank], c)
	}
	switch len(groups) {
	case 2:
		for _, group := range groups {
			if len(group) == 4 {
				return hand{cat: fourKind, rankSum: sum, highestRank: group[0].rank}
			}
			if len(group) == 3 {
				return hand{cat: fullHouse, rankSum: sum, highestRank: group[0].rank}
			}
		}
	case 3:
		hr := 0
		for _, group := range groups {
			if len(group) == 3 {
				return hand{cat: threeKind, rankSum: sum, highestRank: group[0].rank}
			}
			// for two pair, take the higher rank pair to be the highest rank.
			if len(group) == 2 && group[0].rank > hr {
				hr = group[0].rank
			}
		}
		return hand{cat: twoPair, rankSum: sum, highestRank: hr}
	case 4:
		for _, group := range groups {
			if len(group) == 2 {
				return hand{cat: onePair, rankSum: sum, highestRank: group[0].rank}
			}
		}
	default:
		flushHand := isFlush(cards)
		straightHand, r := isStraight(cards)
		switch {
		case flushHand && straightHand:
			return hand{cat: straightFlush, rankSum: sum, highestRank: r}
		case straightHand:
			return hand{cat: straight, rankSum: sum, highestRank: r}
		case flushHand:
			return hand{cat: flush, rankSum: sum, highestRank: getHighestRank(cards)}
		default:
			return hand{cat: highCard, rankSum: sum, highestRank: getHighestRank(cards)}
		}
	}
	panic("invalid hand")
}

// check whether it's straight and if so also returns its highest rank.
func isStraight(cards []card) (bool, int) {
	r1, r2, r3, r4, r5 := cards[0].rank, cards[1].rank, cards[2].rank, cards[3].rank, cards[4].rank
	sr := []int{r1, r2, r3, r4, r5}
	sort.Ints(sr)

	// aces can start a straight which is the smallest straight! Aces effectively is ranked 1 in such case.
	if sr[0] == 2 && sr[1] == 3 && sr[2] == 4 && sr[3] == 5 && sr[4] == 14 {
		return true, 5
	}
	// check for a normal straight, including the one ends with an aces.
	if sr[4]-sr[3] == 1 && sr[3]-sr[2] == 1 && sr[2]-sr[1] == 1 && sr[1]-sr[0] == 1 {
		return true, sr[4]
	}
	return false, 0
}

func isFlush(cards []card) bool {
	s1, s2, s3, s4, s5 := cards[0].suit, cards[1].suit, cards[2].suit, cards[3].suit, cards[4].suit
	return s1 == s2 && s2 == s3 && s3 == s4 && s4 == s5
}
