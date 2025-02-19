import "dotenv/config";
import { listAllServers } from "./game.js";
import { capitalize, InstallGlobalCommands } from "./utils.js";

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
  name: "ck-list",
  description: "List CreeperKeeper Servers",
  options: [
    {
      type: listAllServers.length,
      name: "server",
      required: true,
      choices: createCommandChoices(),
    },
  ],
};

const ALL_COMMANDS = [LIST_SERVERS_COMMAND];

InstallGlobalCommands(process.env.APP_ID, ALL_COMMANDS);
