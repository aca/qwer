use str
var partial-cmd = ["echo2" ""]
var current-input = (str:trim-space (str:join ' ' $partial-cmd))
echo "current-input" $current-input

# Function to find completion candidate
fn find-completion {|full partial|
  var full-words = [(str:split ' ' $full)]
  var partial-words = [(str:split ' ' $partial)]
  
  # Get the last word of partial (the prefix to complete)
  var prefix = $partial-words[-1]
  
  # Get the position where we need to look for completion
  var position = (- (count $partial-words) 1)
  
  # Check if there's a word at that position in full string
  if (< $position (count $full-words)) {
    var candidate = $full-words[$position]
    
    # Check if the candidate starts with the prefix
    if (str:has-prefix $candidate $prefix) {
      echo $candidate
    } else {
      echo ""
    }
  } else {
    echo ""
  }
}

qwer --list | each { |line|
    if (not (str:has-prefix $line $current-input)) {
        continue
    }

    find-completion $line $current-input
    find-completion $line "echo2"
}
