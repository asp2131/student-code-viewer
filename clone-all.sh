filename="list$1.txt"
for i in $(cat < $filename); do
    name=$(echo $i | tr -d " \t\n\r")
    git clone https://github.com/$name/$name.github.io $name
done