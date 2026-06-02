local function terminal(command)
  vim.cmd("botright new")
  vim.fn.termopen(command)
  vim.cmd("startinsert")
end

local commands = {
  BootWranglerTest = { "go", "test", "./..." },
  BootWranglerFrontendCheck = { "npm", "--prefix", "frontend", "run", "lint" },
  BootWranglerFrontendTest = { "npm", "--prefix", "frontend", "run", "test" },
  BootWranglerDesktopBuild = { "./scripts/build-desktop.sh" },
}

for name, command in pairs(commands) do
  vim.api.nvim_create_user_command(name, function()
    terminal(command)
  end, {})
end
