# Routing matrix

## Scoring (0–10 per persona)

For each persona, add points when question matches:

| Signal | jeff | jensen | hormozi | satya | elon | jobs | sundar |
|--------|------|--------|---------|-------|------|------|--------|
| Pricing / offer / conversion | 2 | 0 | **10** | 1 | 1 | 3 | 1 |
| Unit economics / CAC / LTV | 3 | 1 | **10** | 2 | 2 | 1 | 1 |
| Reversible vs irreversible decision | **10** | 4 | 3 | 4 | 5 | 4 | 4 |
| Customer obsession / working backwards | **9** | 3 | 5 | 4 | 2 | **9** | 5 |
| Org speed / flat / bureaucracy | 6 | **10** | 2 | 7 | 6 | 3 | 4 |
| Culture / transformation / mindset | 5 | 4 | 2 | **10** | 2 | 4 | 6 |
| Platform / ecosystem / partners | 7 | 8 | 2 | **9** | 4 | 3 | **8** |
| R&D / hard tech / manufacturing | 4 | **10** | 1 | 5 | **10** | 4 | 6 |
| First principles / cost / delete steps | 5 | 7 | 4 | 3 | **10** | 5 | 3 |
| Product focus / UX / saying no | 5 | 2 | 3 | 3 | 3 | **10** | 5 |
| AI strategy / safety / trust at scale | 4 | 7 | 2 | 8 | 5 | 4 | **10** |
| Build vs buy (software) | 6 | 5 | 2 | 6 | **8** | 4 | 7 |
| Build vs buy (hardware) | 5 | **9** | 1 | 3 | **10** | 3 | 3 |
| Hiring / talent density | 5 | 6 | 4 | 7 | 5 | **9** | 5 |
| Launch now vs polish | **8** | 6 | 4 | 4 | 7 | **8** | 6 |
| Kill project / scope cut | 6 | 4 | 3 | 4 | 5 | **10** | 4 |
| Scale go-to-market | 6 | 2 | **9** | 4 | 2 | 3 | 3 |
| Enterprise / B2B alignment | 6 | 5 | 6 | **9** | 3 | 4 | 7 |

**Bold** = primary owner for that row.

## Roster size rules

| Condition | Pick |
|-----------|------|
| `--count N` | Top N scorers |
| `--solo id` | 1 |
| `--only a,b,c` | Exactly those (max 4) |
| Top score ≥ 8 and second ≥ 6, gap to third ≥ 3 | **2** advisors |
| Top three within 2 points of each other | **3** advisors |
| Default | **2** if narrow, **3** if strategic/multi-domain |

Never pick advisors scoring **< 4** unless `--add` or explicit name in question.

## Complementary pairs (auto-add second when first wins)

When primary selected, consider boosting complementary +2:

| Primary | Complement | Why |
|---------|------------|-----|
| hormozi | jeff | Economics vs decision speed |
| hormozi | jobs | Offer vs experience |
| jeff | jobs | Decision process vs product focus |
| jensen | elon | Org speed vs delete/simplify |
| satya | sundar | Culture vs AI-at-scale |
| jobs | jeff | Focus vs reversible launch |
| elon | jensen | Engineering vs org throughput |
| sundar | satya | AI rollout vs transformation |

## Example routings

**"Should we raise price from $29 to $49 or add enterprise tier?"**
→ hormozi (10), jeff (6), jobs (4) → **hormozi, jeff**

**"Engineering org has 6 layers, projects take 9 months"**
→ jensen (10), satya (7), jeff (6) → **jensen, satya** or **jensen, jeff**

**"Launch AI chat feature to 10M users next month"**
→ sundar (10), satya (8), jeff (6) → **sundar, satya, jeff**

**"Cut roadmap from 12 features to 3"**
→ jobs (10), jeff (6) → **jobs, jeff**

**"Should we vertically integrate our API layer?"**
→ elon (10), jensen (8), jeff (6) → **elon, jensen**

**"Gym-style membership business struggling with churn"**
→ hormozi (10) → **hormozi** alone or **hormozi, jeff**

**"Legacy enterprise software company moving to cloud SaaS"**
→ satya (10), sundar (7), jeff (6) → **satya, sundar**

## Explicit user override phrases

| User says | Action |
|-----------|--------|
| "only Jensen" / "just Jeff" | `--solo` |
| "ask Jeff and Hormozi" | `--only jeff,hormozi` |
| "council + add Elon" | auto-route + `--add elon` |
| "what would Jobs think" (no others) | `--solo jobs` |
| "Jensen, org is slow" (open council) | auto-route ensuring jensen included |