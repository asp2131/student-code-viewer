for i in $(cat < "list.txt"); do
    name=$(echo $i | tr -d " \t\n\r")
    git clone https://github.com/$name/$name.github.io $name
done