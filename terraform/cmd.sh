sudo docker run -d --name mc \
  -p 25565:25565 \
  -e EULA=TRUE -e \
  TYPE=FTBA -e \
  FTB_MODPACK_ID=126 -e \
  FTB_MODPACK_VERSION=100011 -e \
  MEMORY=2G -e \
  MAX_PLAYERS=10 -e \
  MOTD="RedCraft" -e \
  TZ=EST -e \
  DIFFICULTY=3 -e \
  OPS="Oldjimmy_" -e \
  RCON_CMDS_LAST_DISCONNECT="stop" \
  -v "$(pwd)/data:/data" \
  --tty \
  --interactive \
  itzg/minecraft-server
