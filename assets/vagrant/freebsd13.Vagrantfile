Vagrant.configure("2") do |config|
  config.vm.box = "generic/freebsd13"
  config.vm.define :freebsd13
  config.vm.hostname = "freebsd13"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg install --quiet --yes git gnupg go zip
  SHELL
  config.vm.provision "file", source: "assets/vagrant/freebsd13.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
