Vagrant.configure("2") do |config|
  config.vm.box = "generic-x32/debian11"
  config.vm.define "debian11-i386"
  config.vm.hostname = "debian11-i386"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    DEBIAN_FRONTEND=noninteractive apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y age gpg golang unzip zip
  SHELL
  config.vm.provision "file", source: "assets/vagrant/debian11-i386.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
