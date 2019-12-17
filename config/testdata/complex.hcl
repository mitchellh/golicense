allow = ["one",
  "two", "three/four"
]

deny = [
    "five",
]

override = {
   "six" = "seven"
}

translate = {
    "eight" = "nine"
    "/gopkg.in/(.*)/" = "github.com/\\1"
}
