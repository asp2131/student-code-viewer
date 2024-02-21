for i in $(cat < "list.txt"); do
    cd $name
    git stash
    cd ..
done