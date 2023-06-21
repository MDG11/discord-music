package main

import (
	"MDG11/discord-music/youtube"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	currentVoiceState *discordgo.VoiceConnection

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "basic-command",
			Description: "Basic command",
		},
		{
			Name:        "echo",
			Description: "Repeats the message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "Message to repeat",
					Required:    true,
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
		{
			Name:        "join",
			Description: "Join voice channel",
		},
		{
			Name:        "play",
			Description: "Play a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "url",
					Description: "Song url",
					Required:    true,
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
	}

	handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		},
		"echo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: i.ApplicationCommandData().Options[0].Value.(string),
				},
			})
		},
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			joinVoice(s, i)
			dgvoice.PlayAudioFile(currentVoiceState, "./tracks/song.mp3", make(chan bool))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Joined the voice channel!",
				},
			})
		},
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			url := i.ApplicationCommandData().Options[0].Value.(string)

			joinVoice(s, i)
			streamUrl := youtube.GetStreamUrl(url)
			// response, err := http.Get(streamUrl)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			os.Remove("./tracks/song.mp3")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Successfully queued a song!",
				},
			})
			downloadFile(streamUrl, "./tracks/song.mp3")
			log.Println("downloaded")
			dgvoice.PlayAudioFile(currentVoiceState, "./tracks/song.mp3", make(chan bool))
		},
	}
)

func main() {
	godotenv.Load()
	token := os.Getenv("DISCORD_TOKEN")
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	err = sess.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := sess.ApplicationCommandCreate(sess.State.User.ID, *&v.GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer sess.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Shutting down!")
}

func joinVoice(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, _ := s.State.Guild(i.GuildID)

	for _, voice := range guild.VoiceStates {
		if voice.UserID == i.Member.User.ID {
			dgv, err := s.ChannelVoiceJoin(i.GuildID, voice.ChannelID, false, false)
			if err != nil {
				log.Fatalf("Cant join %s", err)
			}
			currentVoiceState = dgv
		}
	}
}

func downloadFile(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
