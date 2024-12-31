package schema

// type voterPoll struct {
// 	PollID   uint
// 	VoteDate time.Time
// }

// type Voter struct {
// 	VoterID     uint
// 	FirstName   string
// 	LastName    string
// 	VoteHistory []voterPoll
// }

type Voter struct {
	VoterID     uint // Change to Link
	FirstName   string
	LastName    string
	VoteHistory []string //  Change to Link
}
