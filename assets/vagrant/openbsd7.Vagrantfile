Vagrant.configure("2") do |config|
  config.vm.box = "generic/openbsd7"
  config.vm.define :openbsd7
  config.vm.hostname = "openbsd7"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg_add -x bzip2 git gnupg go zip
    echo CHEZMOI_GITHUB_ACCESS_TOKEN=#{ENV['CHEZMOI_GITHUB_ACCESS_TOKEN']} >> /home/vagrant/.bash_profile
    echo CHEZMOI_GITHUB_TOKEN=#{ENV['CHEZMOI_GITHUB_TOKEN']} >> /home/vagrant/.bash_profile
    echo GITHUB_ACCESS_TOKEN=#{ENV['GITHUB_ACCESS_TOKEN']} >> /home/vagrant/.bash_profile
    echo GITHUB_TOKEN=#{ENV['GITHUB_TOKEN']} >> /home/vagrant/.bash_profile
    echo export CHEZMOI_GITHUB_ACCESS_TOKEN CHEZMOI_GITHUB_TOKEN GITHUB_ACCESS_TOKEN GITHUB_TOKEN >> /home/vagrant/.bash_profile
  SHELL
  config.vm.provision "file", source: "assets/vagrant/openbsd7.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
