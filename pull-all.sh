filename="list$1.txt"
BOLD='\033[1m'
NONE='\033[00m'

# Initialize arrays to hold usernames
committed=()
not_committed=()
no_repo=()

# Loop through each username in the file
for i in $(cat < $filename); do
    if test -e "$i"; then
        cd $i
        git pull --quiet

        # Check if there has been a commit within the last hour
        git_status=$(git log --since="1 hours ago" --pretty=format:"%h" -1)
        
        if test -z "$git_status"; then
            not_committed+=("$i")
        else
            committed+=("$i")
        fi
        cd ..
    else
        no_repo+=("$i")
    fi
done

# Function to print a list in a box
print_box() {
    local title="$1"
    shift
    local items=("$@")

    echo "\t-------------------------------------------------------------------------------"
    printf "\t| %-69s |\n" "$title"
    echo "\t-------------------------------------------------------------------------------"
    for item in "${items[@]}"; do
        printf "\t| %-75s |\n" "$item"
    done
    echo "\t-------------------------------------------------------------------------------"
    echo ""
}

# Print the results in boxes
echo ""
echo ""

if [ ${#committed[@]} -gt 0 ]; then
    print_box "✅ ✅ ✅ These students have pushed code up in the last hour ✅ ✅ ✅" "${committed[@]}"
else
    echo "    ❗❗No students have pushed work up in the last hour❗❗"
fi
echo ""

if [ ${#not_committed[@]} -gt 0 ]; then
    print_box "🚧 🚧 🚧 These students have NOT pushed code up in the last hour 🚧 🚧 🚧" "${not_committed[@]}"
fi
echo ""

if [ ${#no_repo[@]} -gt 0 ]; then
    print_box "🚩 🚩 🚩 These students do not have a correctly configured repo 🚩 🚩 🚩" "${no_repo[@]}"
fi
echo ""