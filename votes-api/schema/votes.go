package schema

type Vote struct {
	VoteID    uint
	VoterID   string
	PollID    string
	VoteValue uint
}
