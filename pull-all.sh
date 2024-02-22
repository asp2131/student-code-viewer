filename="list$1.txt"
BOLD='\033[1m'
NONE='\033[00m'
for i in $(cat < $filename); do
    if test -e "$i"; then
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
echo -e "${BOLD}commits within the last hour are below:"
echo ""

for i in $(cat < $filename); do
    if test -e "$i"; then
        name=$(echo $i | tr -d " \t\n\r")
        tput bold
        echo $name
        tput sgr0
        cd $name
        PAGER="/bin/cat"
        status=$(git log --since="1 hours ago" --date=format-local:'%a, %b %d %H:%M:%S' --pretty=format:"✅ ✅ ✅ $name has committed today! committed at %cd ✅ ✅ ✅" -1)
        if test -z "$status"; then
            echo "🚧 🚧 🚧 $name has not committed within the last hour! 🚧 🚧 🚧"
            # echo -e "🚧 🚧 🚧 $name has${BOLD} not${NONE} committed within the last hour! 🚧 🚧 🚧"
        else
            echo $status
        fi
        cd ..
    else
        echo "🚩 🚩 🚩 $i does not have a github repo that matches the expected name. Check with them to ensure they have named their repository correctly 🚩 🚩 🚩"
        echo "🚩 🚩 🚩 Visit https://github.com/$i to see their existing repositories 🚩 🚩 🚩"
    fi
    echo ""
done