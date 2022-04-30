Vagrant.configure("2") do |config|
  config.vm.box = "generic/openbsd7"
  config.vm.define :openbsd7
  config.vm.hostname = "openbsd7"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg_add -x bzip2 git gnupg go zip
  SHELL
  config.vm.provision "file", source: "assets/vagrant/openbsd7.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
