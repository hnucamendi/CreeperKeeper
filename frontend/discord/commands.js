import "dotenv/config";
import { listAllServers } from "./game.js";
import { capitalize, InstallGlobalCommands } from "./utils.js";
const { SlashCommandBuilder } = require("@discordjs/builders");

// Get the game choices from game.js
async function createCommandChoices() {
  const choices = await listAllServers();
  const commandChoices = [];

  for (let choice of choices) {
    commandChoices.push({
      name: capitalize(choice.serverID),
      value: choice.serverName,
    });
  }

  return commandChoices;
}

const LIST_SERVERS_COMMAND = {
  name: "cklist",
  description: "List CreeperKeeper Servers",
  options: [
    {
      type: 3,
      name: "server",
      required: true,
      choices: createCommandChoices(),
    },
  ],
};


module.exports = {
  data: new SlashCommandBuilder()
    .setName("login")
    .setDescription("Get a login link to authenticate with CreeperKeeper."),
  async execute(interaction) {
    const auth0Domain = "YOUR_AUTH0_DOMAIN";
    const clientId = "YOUR_AUTH0_CLIENT_ID";
    const redirectUri = encodeURIComponent(
      "https://your-backend.com/auth/callback",
    );

    const loginUrl = `https://${auth0Domain}/authorize?response_type=code&client_id=${clientId}&redirect_uri=${redirectUri}&scope=openid profile email`;

    await interaction.reply(`Click here to log in: [Login](${loginUrl})`);
  },
};
const ALL_COMMANDS = [LIST_SERVERS_COMMAND];

InstallGlobalCommands(process.env.APP_ID, ALL_COMMANDS);
