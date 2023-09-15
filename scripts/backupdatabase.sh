_now=$(date +"%Hh-%Y_%m_%d")
_file="./backup/nanoymousdb_$_now.dump"

pg_dump -U postgres -Fc nanonymousdb > "$_file"

# Delete everything older than a month but still keep one a month
find ./backup/* -type f -not -path './backup/*01.dump' -mtime +30 -exec rm {} \;
