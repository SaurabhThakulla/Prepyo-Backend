# =============================================================================
# Prepyo end-to-end smoke test
# =============================================================================
# Registers a throwaway account against the running stack, then exercises the
# real endpoints with that session. Requires both servers to be up:
#   npm run dev:api    (http://localhost:8080)
#   npm run dev        (http://localhost:3000)
#
# Everything goes through :3000 so the Next proxy is covered too.
# =============================================================================

$base = "http://localhost:3000"
$pass = 0
$fail = 0

function Check($label, $ok, $detail = "") {
  if ($ok) {
    Write-Host "[PASS] $label $detail" -ForegroundColor Green
    $script:pass++
  } else {
    Write-Host "[FAIL] $label $detail" -ForegroundColor Red
    $script:fail++
  }
}

Write-Host "========================================="
Write-Host "PREPYO END-TO-END SMOKE TEST"
Write-Host "========================================="

# --- Pages render (signed out) -----------------------------------------------
Write-Host "`nPublic pages"
foreach ($path in @("/", "/login", "/signup")) {
  try {
    $r = Invoke-WebRequest -Uri "$base$path" -UseBasicParsing -TimeoutSec 10
    Check "GET $path" ($r.StatusCode -eq 200) "-> $($r.StatusCode)"
  } catch {
    Check "GET $path" $false "-> $($_.Exception.Message)"
  }
}

# --- Protected routes bounce signed-out visitors ------------------------------
Write-Host "`nRoute protection"
try {
  $r = Invoke-WebRequest -Uri "$base/dashboard" -UseBasicParsing -MaximumRedirection 0 -ErrorAction SilentlyContinue
  Check "GET /dashboard redirects when signed out" ($r.StatusCode -eq 307 -or $r.StatusCode -eq 302) "-> $($r.StatusCode)"
} catch {
  # PowerShell throws on 3xx when MaximumRedirection is 0, which is the pass case.
  Check "GET /dashboard redirects when signed out" $true "-> redirect"
}

# --- Public API ---------------------------------------------------------------
Write-Host "`nPublic API"
try {
  $plans = Invoke-RestMethod -Uri "$base/api/v1/subscriptions/plans" -TimeoutSec 10
  Check "GET /subscriptions/plans" ($plans.success -and $plans.plans.Count -gt 0) "-> $($plans.plans.Count) plans"
} catch {
  Check "GET /subscriptions/plans" $false "-> $($_.Exception.Message)"
}

# --- Register a throwaway account ---------------------------------------------
Write-Host "`nAuthentication"
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$email = "smoke-$([guid]::NewGuid().ToString('N').Substring(0,8))@prepyo.test"
$password = "SmokeTestPassword123"

$registerBody = @{
  email       = $email
  password    = $password
  name        = "Smoke Test"
  nepalRegion = "Kathmandu"
  timezone    = "Asia/Kathmandu"
} | ConvertTo-Json

try {
  $reg = Invoke-RestMethod -Uri "$base/api/v1/auth/register" -Method Post -Body $registerBody `
    -ContentType "application/json" -WebSession $session -TimeoutSec 15
  Check "POST /auth/register" ($reg.success -and $reg.user.email -eq $email) "-> $($reg.user.email)"
} catch {
  Check "POST /auth/register" $false "-> $($_.Exception.Message)"
  Write-Host "`nCannot continue without a session. Is the Go API running on :8080?" -ForegroundColor Yellow
  exit 1
}

# Wrong password must be rejected.
$badBody = @{ email = $email; password = "definitely-not-it" } | ConvertTo-Json
try {
  Invoke-RestMethod -Uri "$base/api/v1/auth/login" -Method Post -Body $badBody `
    -ContentType "application/json" -TimeoutSec 10 | Out-Null
  Check "POST /auth/login rejects a wrong password" $false "-> unexpectedly succeeded"
} catch {
  Check "POST /auth/login rejects a wrong password" $true "-> 401"
}

# Correct password must work.
$goodBody = @{ email = $email; password = $password } | ConvertTo-Json
try {
  $login = Invoke-RestMethod -Uri "$base/api/v1/auth/login" -Method Post -Body $goodBody `
    -ContentType "application/json" -WebSession $session -TimeoutSec 15
  Check "POST /auth/login" ($login.success) "-> $($login.user.name)"
} catch {
  Check "POST /auth/login" $false "-> $($_.Exception.Message)"
}

# --- Authenticated reads -------------------------------------------------------
Write-Host "`nAuthenticated API"
$reads = @(
  @{ path = "/api/v1/profile";           key = "user" },
  @{ path = "/api/v1/progress";          key = "skills" },
  @{ path = "/api/v1/gamification";      key = "dailyMissions" },
  @{ path = "/api/v1/questions?exam=PTE"; key = "questions" },
  @{ path = "/api/v1/mocks?exam=PTE";    key = "mocks" },
  @{ path = "/api/v1/mistakes";          key = "mistakes" },
  @{ path = "/api/v1/leaderboards";      key = "leaderboard" },
  @{ path = "/api/v1/notifications";     key = "notifications" },
  @{ path = "/api/v1/referrals/me";      key = "referralCode" },
  @{ path = "/api/v1/subscriptions";     key = "subscription" }
)

foreach ($r in $reads) {
  try {
    $res = Invoke-RestMethod -Uri "$base$($r.path)" -WebSession $session -TimeoutSec 15
    Check "GET $($r.path)" ($res.success -and $null -ne $res.($r.key))
  } catch {
    Check "GET $($r.path)" $false "-> $($_.Exception.Message)"
  }
}

# --- Practice submission is graded server-side ---------------------------------
Write-Host "`nGrading"
try {
  $questions = Invoke-RestMethod -Uri "$base/api/v1/questions?exam=PTE&skill=reading" -WebSession $session -TimeoutSec 15
  $q = $questions.questions | Select-Object -First 1

  if ($null -eq $q) {
    Check "Practice submission" $false "-> no reading questions seeded"
  } else {
    $answer = @{ questionId = $q.id; timeSpentSeconds = 30 }
    if ($q.blanks) {
      $blanks = @{}
      foreach ($b in $q.blanks) { $blanks[$b.id] = ($b.options | Select-Object -First 1) }
      $answer.blankResponses = $blanks
    }

    $submitted = Invoke-RestMethod -Uri "$base/api/v1/practice/attempts" -Method Post `
      -Body ($answer | ConvertTo-Json -Depth 5) -ContentType "application/json" `
      -WebSession $session -TimeoutSec 20

    Check "POST /practice/attempts" ($submitted.success -and $null -ne $submitted.attempt) `
      "-> scored $($submitted.attempt.score)/$($submitted.attempt.maxScore), +$($submitted.xpAwarded) XP"
  }
} catch {
  Check "POST /practice/attempts" $false "-> $($_.Exception.Message)"
}

# --- Sign out -------------------------------------------------------------------
Write-Host "`nSession teardown"
try {
  Invoke-RestMethod -Uri "$base/api/v1/auth/logout" -Method Post -WebSession $session -TimeoutSec 10 | Out-Null
  try {
    Invoke-RestMethod -Uri "$base/api/v1/profile" -WebSession $session -TimeoutSec 10 | Out-Null
    Check "Session is revoked after logout" $false "-> profile still readable"
  } catch {
    Check "Session is revoked after logout" $true "-> 401"
  }
} catch {
  Check "POST /auth/logout" $false "-> $($_.Exception.Message)"
}

Write-Host "`n-----------------------------------------"
Write-Host "Results: $pass passed, $fail failed"
Write-Host "Test account left behind: $email"
if ($fail -gt 0) { exit 1 }
