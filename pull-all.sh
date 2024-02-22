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
        tput bold
        echo $name
        tput sgr0
        cd $name
        PAGER="/bin/cat"
        status=$(git log --since="1 hour ago" --pretty=format-local:"✅ ✅ ✅ $name has committed today! committed at %m-%d %H:%M:%S ✅ ✅ ✅" -1)
        if test -z $status; then
            echo "🚧 🚧 🚧 $name has not committed within the last hour! 🚧 🚧 🚧"
        else
            echo $status
        fi
        cd ..
    else
        echo "🚩 🚩 🚩 $i does not have a github repo that matches the expected name. Check with them to ensure they have named their repository correctly."
        echo "Visit https://github.com/$i to see their existing repositories 🚩 🚩 🚩"
    fi
    echo ""
done