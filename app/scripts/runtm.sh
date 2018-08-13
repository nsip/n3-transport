export TMHOME=~/tendermint-n3

echo "TMHOME is:" $TMHOME

echo "deleting old data"
./tendermint unsafe_reset_all
echo "old data deleted"
echo "starting b/c node..."
./tendermint node --proxy_app=n3 --consensus.create_empty_blocks=false
# ./tendermint node --consensus.create_empty_blocks=false
