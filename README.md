*pascal* (/pæˈskæl/) named after **Blaise Pascal**, completing the duo with *blaise*. While *blaise* handles the heavy lifting of routing and pathfinding, *pascal* provides the human connection, acting as a smart interface for your daily commute.

*pascal* is a proactive, chat-based transit assistant. It integrates directly into your messaging apps (like Discord) to manage recurring trips, monitor your routes in real-time, and notify you exactly when it's time to leave, so you never have to check a timetable again.

## Why pascal?

Most transit apps are passive; you have to open them, search, and stare at the screen. pascal flips this model by being an active assistant that lives where you already chat:

- **Set and Forget**: Define your routine once using natural language ("Work, 9 AM every weekday"), and *pascal* handles the rest.

- **Proactive Alerts**: Instead of you checking the time, *pascal* tells you: "Leave in 15 minutes" or "Stop coming up in 1 minute."

- **Blaise Powered**: Built on top of the *blaise* engine, ensuring that every route calculation is blazing fast, private, and local.

- **Self-Hosted SaaS**: Run it for your friends, your server, or your organization. You own the data and the infrastructure.

## Core Features

- **Natural Language Scheduling**: Create complex recurring trips effortlessly.
    "Work, from: Home to: Office every day, arrival 08:00:00 except saturday and sunday"

- **Real-Time Guidance**: Continuous monitoring of your active trips with push notifications for departures, transfers, and arrivals.


[!NOTE] This project is both a SaaS solution and an experiment in human-centric interfaces for the *blaise* engine.


## Installation

*pascal* is written in Go. You can build it from source:

```bash
git clone https://github.com/vincbro/pascal.git
cd pascal
go build -o pascal ./cmd/bot
```

## Quick Start
1. **Prerequisites**: Ensure you have a running instance of the blaise server and a Discord Bot Token.

2. **Configuration**: Create a .env file in the root directory:
```env
DISCORD_KEY=your_discord_bot_token
APP_ID=your_discord_app_id
GUILD_ID=your_discord_guild_id
BLAISE_URL=http://localhost:8080    
```

3. **Run**
```bash
./pascal
```

4. **Usage:** In your Discord server, use the slash command to find a route:
```txt
/new from: "Central Station" to: "Tech Park"
```


## Roadmap

- [x] Discord Bot integration with Slash Commands
- [x] Basic routing via blaise client
- [ ] Natural Language Processing (NLP) for trip scheduling
- [ ] Job scheduler for recurring trip monitoring
- [ ] Real-time state tracking and push notifications
- [ ] Integration with other platforms (Slack, Telegram)


## References
- [Blaise Engine](https://github.com/vincbro/blaise)
- [DiscordGo](https://github.com/bwmarrin/discordgo)

## License
This project is licensed under the MIT License.
