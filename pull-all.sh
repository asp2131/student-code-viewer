for i in $(cat < "list.txt"); do
    name=$(echo $i | tr -d " \t\n\r")
    echo $name
    cd $name
    git pull --ff
    cd ..
done

echo "\n"
echo "---------"
echo "\n"
echo "commits within the last hour are below:"

for i in $(cat < "list.txt"); do
    name=$(echo $i | tr -d " \t\n\r")
    echo $name
    cd $name
    PAGER="/bin/cat"
    git log --since="1 hour ago" --pretty=format:"^^^^^ committed at %cd" -1
    echo "\n"
    cd ..
done