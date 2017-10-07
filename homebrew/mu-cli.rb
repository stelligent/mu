# Some comments are needed for the Makefile to substitute lines, do not remove!
class MuCli < Formula
  desc "A full-stack DevOps on AWS tool"
  homepage "https://github.com/stelligent/mu"
  if OS.mac?
    url "https://github.com/stelligent/mu/releases/download/v1.0.1/mu-darwin-amd64" # The MacOS master url
    sha256 "ccd937d51b52a8727271230d601e296797d645574c55eae8f98aef12a8f0c249" # The MacOS master sha256sum
  elsif OS.linux?
    url "https://github.com/stelligent/mu/releases/download/v1.0.1/mu-linux-amd64" # The Linux master url
    sha256 "6679e14f3138578429a115b88c9b8eab1e411aafc3a4b5c542bd2b8f5f1ab4f5" # The Linux master sha256sum
  end
  version "v1.0.1" # The master version
  devel do
    if OS.mac?
    url "https://github.com/stelligent/mu/releases/download/v1.1.1-develop/mu-darwin-amd64" # The MacOS develop url
    sha256 "f313ba935758817da792f51d31456e2df0f99dc28c8dcbbc1dac042f66c16578" # The MacOS develop sha256sum
    elsif OS.linux?
    url "https://github.com/stelligent/mu/releases/download/v1.1.1-develop/mu-linux-amd64" # The Linux develop url
    sha256 "48302e85dd0a516220f4d5492007657c487a95daee471ee36866bb6cf47d6531" # The Linux develop sha256sum
    end
    version "v1.1.1-develop" # The develop version
  end

  bottle :unneeded

  def install
    if OS.mac?
      bin.install "mu-darwin-amd64"
      mv "#{bin}/mu-darwin-amd64", "#{bin}/mu-cli"
    elsif OS.linux?
      bin.install "mu-linux-amd64"
      mv "#{bin}/mu-linux-amd64", "#{bin}/mu-cli"
    end
  end

  test do
    system "#{bin}/mu-cli" ,"--version"
  end
end
