Vagrant.configure("2") do |config|
  config.vm.box = "generic-x32/debian11"
  config.vm.define "debian11-i386"
  config.vm.hostname = "debian11-i386"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    DEBIAN_FRONTEND=noninteractive apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y age gpg golang unzip xz-utils zip
  SHELL
  config.vm.provision "shell", inline: <<-SHELL
    echo CHEZMOI_GITHUB_ACCESS_TOKEN=#{ENV['CHEZMOI_GITHUB_ACCESS_TOKEN']} >> /home/vagrant/.bash_profile
    echo CHEZMOI_GITHUB_TOKEN=#{ENV['CHEZMOI_GITHUB_TOKEN']} >> /home/vagrant/.bash_profile
    echo GITHUB_ACCESS_TOKEN=#{ENV['GITHUB_ACCESS_TOKEN']} >> /home/vagrant/.bash_profile
    echo GITHUB_TOKEN=#{ENV['GITHUB_TOKEN']} >> /home/vagrant/.bash_profile
    echo export CHEZMOI_GITHUB_ACCESS_TOKEN CHEZMOI_GITHUB_TOKEN GITHUB_ACCESS_TOKEN GITHUB_TOKEN >> /home/vagrant/.bash_profile
  SHELL
  config.vm.provision "file", source: "assets/vagrant/debian11-i386.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
