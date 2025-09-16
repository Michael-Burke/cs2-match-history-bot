package internal

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var responded sync.Map

func RegisterSlashCommands(s *discordgo.Session) {
	permManageGuild := int64(discordgo.PermissionManageServer)
	dmDisabled := false

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "refresh",
			Description: "Refreshes the current and last week's FACEIT statistics for all listed players",
		},
		{
			Name:                     "list-players",
			Description:              "Lists all players currently being tracked",
			DefaultMemberPermissions: &permManageGuild,
			DMPermission:             &dmDisabled,
		},
		{
			Name:                     "add-player",
			Description:              "Adds a player to the list of players being tracked",
			DefaultMemberPermissions: &permManageGuild,
			DMPermission:             &dmDisabled,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The name of the player to ADD",
					Required:    true,
				},
			},
		},
		{
			Name:                     "remove-player",
			Description:              "Removes a player from the list of players being tracked",
			DefaultMemberPermissions: &permManageGuild,
			DMPermission:             &dmDisabled,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The name of the player to REMOVE",
					Required:    true,
				},
			},
		},
	}
	registeredCmds := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(applicationID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCmds[i] = cmd
		log.Println("Slash command registered:", v.Name)
	}
	log.Println("Slash commands registered:", len(registeredCmds))
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"refresh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			content := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "refreshing...",
				},
			})
			go func() {
				content = FACEITInit(s)
				if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				}); err != nil {
					log.Printf("failed to edit response: %v", err)
				}
			}()
		},
		"list-players": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !hasManageGuildPermission(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "You do not have permission to use this command.",
					},
				})
				return
			}
			content := ""
			content = ListPlayers()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: content,
				},
			})
		},
		"add-player": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !hasManageGuildPermission(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "You do not have permission to use this command.",
					},
				})
				return
			}
			content := ""
			content = AddPlayer(i.ApplicationCommandData().Options[0].StringValue())
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: content,
				},
			})
		},
		"remove-player": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !hasManageGuildPermission(i) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "You do not have permission to use this command.",
					},
				})
				return
			}
			content := ""
			content = RemovePlayer(i.ApplicationCommandData().Options[0].StringValue())
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: content,
				},
			})
		},
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

// hasManageGuildPermission returns true if the invoking member has Administrator or Manage Guild
func hasManageGuildPermission(i *discordgo.InteractionCreate) bool {
	if i == nil || i.Member == nil {
		return false
	}
	perms := i.Member.Permissions
	if perms&discordgo.PermissionAdministrator != 0 {
		return true
	}
	if perms&discordgo.PermissionManageServer != 0 {
		return true
	}
	return false
}

// UPDATE the BOT Presence Value
func UpdatePresence(s *discordgo.Session, discordMessage string, marker string) error {
	activity := "Watching"

	if updateChannelID != "" {
		msg := fmt.Sprintf("%s", discordMessage)
		if messageID, err := getStatusMessageID(s, updateChannelID, marker); err == nil && messageID != "" {
			editStatusMessage(s, updateChannelID, messageID, msg)
		} else {
			postMessage(s, updateChannelID, msg)
		}
		msg = ""
	}
	return s.UpdateGameStatus(0, fmt.Sprintf("%s: %s", gameName, activity))
}

func UpdateMessage(s *discordgo.Session, discordMessage string, marker string) {
	if updateChannelID != "" {
		msg := fmt.Sprintf("%s", discordMessage)
		if messageID, err := getStatusMessageID(s, updateChannelID, marker); err == nil && messageID != "" {
			editStatusMessage(s, updateChannelID, messageID, msg)
		} else {
			postMessage(s, updateChannelID, msg)
		}
		msg = ""
	}
}

// Post Message to Discord
// Example: postMessage(s, updateChannelID, msg)
// Needs: s *discordgo.Session, channelID string, message string
func postMessage(s *discordgo.Session, channelID string, message string) error {
	// log.Printf("Posting message to discord channel ID %s: \n  - %s", channelID, message)
	ch, err := s.Channel(channelID)
	if err != nil {
		log.Println("Error fetching channel", channelID, err)
		return err
	}
	switch ch.Type {
	case discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews,
		discordgo.ChannelTypeGuildPublicThread, discordgo.ChannelTypeGuildPrivateThread, discordgo.ChannelTypeGuildNewsThread:
		// ok to send
	default:
		return fmt.Errorf("unsupported channel type %d for sending messages; use a text channel or thread", ch.Type)
	}
	_, err = s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Println("Error posting message to discord channel", err)
	}
	return err
}

func getStatusMessageID(s *discordgo.Session, channelID string, marker string) (string, error) {
	msgs, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return "", err
	}
	for _, msg := range msgs {
		if msg.Author != nil && msg.Author.ID == s.State.User.ID && strings.Contains(msg.Content, marker) {
			return msg.ID, nil
		}
	}
	return "", nil
}

func editStatusMessage(s *discordgo.Session, channelID string, messageID string, message string) error {
	// log.Println("Editing message in discord channel", channelID, message)
	_, err := s.ChannelMessageEdit(channelID, messageID, message)
	if err != nil {
		log.Println("Error editing message in discord channel", err)
	}
	return err
}

func PostUsageMessage(s *discordgo.Session) {
	marker := "**Usage**: "
	content := "\n`/refresh`" + ` to refresh the current and last week's FACEIT statistics for all listed players
` + "`/list-players`" + ` to list all players currently being tracked
` + "`/add-player`" + ` to add a player to the list of players being tracked
` + "`/remove-player`" + ` to remove a player from the list of players being tracked
	`
	// UpdatePresence(s, marker+content+"\n ---- \n", marker)
	UpdateMessage(s, marker+content+"\n ---- \n", marker)
}

func BotInit(s *discordgo.Session) {
	loadEnv(true)
	RegisterSlashCommands(s)
	PostUsageMessage(s)
}
