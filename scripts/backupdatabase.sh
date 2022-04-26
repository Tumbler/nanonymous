_now=$(date +"%Y_%m_%d")
_file="./backup/nanoymousdb_$_now.dump"

pg_dump -U postgres -Fc nanonymousdb > "$_file"
