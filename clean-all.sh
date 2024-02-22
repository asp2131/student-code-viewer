filename="list$1.txt"
for i in $(cat < $filename); do
    if test -e $i; then
        cd $i
        git checkout .
        cd ..
    fi
done