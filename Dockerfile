services:
  mc:
    image: itzg/minecraft-server:latest
    tty: true
    stdin_open: true
    ports:
      - "25565:25565"
    environment:
      EULA: "TRUE"
      TYPE: "FTBA"
      FTB_MODPACK_ID: "126"
      FTB_MODPACK_VERSION_ID: "100011"
      MEMORY: "4G"
      MAX_PLAYERS: "15"
      MOTD: "RedCraft"
      TZ: "America/New_York"
    volumes:
      - "./data:/data"
