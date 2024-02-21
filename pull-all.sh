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
echo "------------------"
echo ""
echo "❗❗❗❗❗❗❗❗❗❗❗❗❗❗❗"
echo ""
# set text to bold
tput bold
echo "commits within the last hour are below:"
# return to normal text
tput sgr0
echo ""

for i in $(cat < "list.txt"); do
    if test -e $i; then
        name=$(echo $i | tr -d " \t\n\r")
        echo $name
        cd $name
        PAGER="/bin/cat"
        git log --since="1 hour ago" --pretty=format:"^^^^^ committed at %cd" -1
        echo ""
        cd ..
    else
        echo "$i does not have a matching github repo. Check with them to ensure they have named their repository correctly. Visit https://github.com/$i to see their existing repositories"
    fi
    echo ""
done