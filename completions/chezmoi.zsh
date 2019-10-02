#compdef chezmoi

_arguments \
  '1: :->level1' \
  '2: :->level2' \
  '3: :->level3' \
  '4: :_files'
case $state in
  level1)
    case $words[1] in
      chezmoi)
        _arguments '1: :(add apply archive cat cd chattr completion data diff docs doctor dump edit edit-config forget help import init merge remove secret source source-path unmanaged update upgrade verify)'
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  level2)
    case $words[2] in
      secret)
        _arguments '2: :(bitwarden generic gopass keepassxc keyring lastpass onepassword pass vault)'
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  level3)
    case $words[3] in
      keyring)
        _arguments '3: :(get set)'
      ;;
      *)
        _arguments '*: :_files'
      ;;
    esac
  ;;
  *)
    _arguments '*: :_files'
  ;;
esac
