for i in $(cat < "list.txt"); do
    if test -e $i; then
        name=$(echo $i | tr -d " \t\n\r")
        echo $name
        cd $name
        git pull --ff
        echo ""
        cd ..
    fi
done

echo ""
echo "---------"
echo "❗❗❗❗❗❗❗❗❗❗❗❗❗❗❗"
echo "commits within the last hour are below:"

for i in $(cat < "list.txt"); do
    if test -e $i; then
        name=$(echo $i | tr -d " \t\n\r")
        echo $name
        cd $name
        PAGER="/bin/cat"
        git log --since="1 hour ago" --pretty=format:"^^^^^ committed at %cd" -1
        echo ""
        echo ""
        cd ..
    fi
done