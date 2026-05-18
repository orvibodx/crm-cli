const { platform, arch } = process
const https = require('https')
const http = require('http')
const fs = require('fs')
const path = require('path')
const zlib = require('zlib')
const { execSync } = require('child_process')

const GOOS = { darwin: 'darwin', linux: 'linux', win32: 'windows' }[platform]
const GOARCH = { x64: 'amd64', arm64: 'arm64' }[arch]

if (!GOOS || !GOARCH) {
  console.error(`Unsupported platform: ${platform} ${arch}`)
  process.exit(1)
}

const pkg = require('./package.json')
const version = pkg.version
const ext = platform === 'win32' ? '.zip' : '.tar.gz'
const archiveName = `crm-cli_${version}_${GOOS}_${GOARCH}${ext}`
const url = `https://github.com/orvibodx/crm-cli/releases/download/v${version}/${archiveName}`

const binDir = path.join(__dirname, 'bin')
const binaryName = platform === 'win32' ? 'crm-cli.exe' : 'crm-cli'
const binPath = path.join(binDir, binaryName)

if (fs.existsSync(binPath)) {
  console.log('crm-cli binary already exists, skipping download.')
  process.exit(0)
}

function followRedirects(url, callback) {
  const client = url.startsWith('https') ? https : http
  client.get(url, { headers: { 'User-Agent': 'node' } }, (res) => {
    if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
      followRedirects(res.headers.location, callback)
    } else {
      callback(res)
    }
  }).on('error', callback)
}

console.log(`Downloading crm-cli v${version} for ${GOOS}/${GOARCH}...`)

followRedirects(url, (res) => {
  if (res.statusCode !== 200) {
    console.error(`Download failed: HTTP ${res.statusCode}`)
    console.error(`URL: ${url}`)
    console.error('')
    console.error('This usually means the release binary has not been published yet.')
    console.error('Build and upload it first: goreleaser release')
    process.exit(1)
  }

  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true })
  }

  const tmpFile = path.join(binDir, archiveName)
  const ws = fs.createWriteStream(tmpFile)
  res.pipe(ws)

  ws.on('finish', () => {
    if (platform === 'win32') {
      execSync(`powershell -command "Expand-Archive -Path '${tmpFile}' -DestinationPath '${binDir}' -Force"`, { stdio: 'inherit' })
    } else {
      execSync(`tar -xzf "${tmpFile}" -C "${binDir}"`, { stdio: 'inherit' })
    }
    fs.unlinkSync(tmpFile)
    finish()
  })
})

function finish() {
  if (!fs.existsSync(binPath)) {
    console.error(`Binary not found at ${binPath}`)
    process.exit(1)
  }
  fs.chmodSync(binPath, 0o755)
  console.log(`crm-cli v${version} installed successfully.`)
}
