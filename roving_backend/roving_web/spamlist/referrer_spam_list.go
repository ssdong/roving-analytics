package spamlist

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"roving_web/config"
	"sync"
)

// referrer_spam_list.txt is a community-curated list of referrer spam domains
// https://github.com/matomo-org/referrer-spam-list/blob/master/spammers.txt

type SpamList struct {
	spamSet map[string]struct{}
}

var (
	instance *SpamList
	once     sync.Once
)

func GetInstance() *SpamList {
	once.Do(func() {
        instance = &SpamList{
            spamSet: make(map[string]struct{}),
        }
        instance.loadSpamList()
		fmt.Println("Loading spam list completed")
    })
	return instance
}

func (s *SpamList) loadSpamList() {
	file, err := os.Open(config.AppConfig.ReferrerSpamFilePath)
	if err != nil {
		log.Fatalf("Failed to open referrer spam list file: %s", err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		s.spamSet[domain] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading referrer spam list file: %s", err.Error())
	}
}

func (s *SpamList) IsSpam(domain string) bool {
	_, exists := s.spamSet[domain]
	return exists
}
