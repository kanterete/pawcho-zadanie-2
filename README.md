# PAwChO – Zadanie 2

## Opis zadania

Rozwiązanie zadania nr 2 z przedmiotu **Programowanie Aplikacji w Chmurze Obliczeniowej**.

Celem zadania było przygotowanie łańcucha **GitHub Actions**, który:

- buduje obraz kontenera na podstawie pliku `Dockerfile` oraz kodu źródłowego aplikacji,
- wspiera dwie architektury: `linux/amd64` oraz `linux/arm64`,
- wykorzystuje dane cache przechowywane w publicznym repozytorium DockerHub,
- wykonuje test CVE obrazu,
- publikuje obraz do GitHub Container Registry tylko wtedy, gdy skan bezpieczeństwa nie wykryje podatności o poziomie `HIGH` lub `CRITICAL`.

## GitHub Actions

Pipeline został zdefiniowany w pliku:

```text
.github/workflows/docker.yml
```

Workflow uruchamia się automatycznie po wykonaniu `push` na gałąź `main` oraz może zostać uruchomiony ręcznie dzięki `workflow_dispatch`.

```yaml
on:
  push:
    branches:
      - main
  workflow_dispatch:
```

Do publikacji obrazu w GitHub Container Registry wykorzystano uprawnienia:

```yaml
permissions:
  contents: read
  packages: write
```

## Etapy działania pipeline

Pipeline składa się z dwóch głównych jobów:

### 1. `scan`

Pierwszy job buduje i skanuje obraz osobno dla dwóch architektur:

```text
linux/amd64
linux/arm64
```

Do tego wykorzystano macierz `matrix`, dzięki której te same kroki są wykonywane dla obu platform.

W tym etapie wykonywane są:

- pobranie kodu z repozytorium,
- konfiguracja QEMU,
- konfiguracja Docker Buildx,
- logowanie do DockerHub,
- zbudowanie obrazu testowego,
- skan CVE za pomocą Trivy.

### 2. `publish`

Drugi job publikuje obraz do GitHub Container Registry. Jest on uruchamiany dopiero po poprawnym zakończeniu joba `scan`.

```yaml
needs: scan
```

Dzięki temu obraz zostaje opublikowany tylko wtedy, gdy skan CVE zakończy się powodzeniem.

## Test CVE

Do sprawdzania podatności wykorzystano skaner **Trivy**.

Wybrano Trivy, ponieważ jest prosty do użycia w GitHub Actions, posiada gotową akcję `aquasecurity/trivy-action` i pozwala łatwo przerwać workflow, gdy zostaną wykryte podatności o określonym poziomie ważności.

W workflow ustawiono:

```yaml
severity: HIGH,CRITICAL
exit-code: 1
vuln-type: os,library
```

Oznacza to, że pipeline zakończy się błędem, jeżeli obraz będzie zawierał podatności sklasyfikowane jako `HIGH` lub `CRITICAL`.

Źródło:

```text
https://github.com/aquasecurity/trivy-action
https://trivy.dev/latest/docs/
```

## Publikacja obrazu

Finalny obraz jest publikowany w GitHub Container Registry:

```text
ghcr.io/kanterete/pawcho-zadanie-2
```

Obraz można pobrać poleceniem:

```bash
docker pull ghcr.io/kanterete/pawcho-zadanie-2:latest
```

## Schemat tagowania obrazów

Dla obrazów publikowanych w GHCR przyjęto dwa typy tagów:

```text
latest
sha-<commit_sha>
```

Przykład:

```text
latest
sha-98c64107dd8c5b3eadd4e2463f6b7a043b2dc0d2
```

Tag `latest` wskazuje najnowszą poprawnie zbudowaną i przeskanowaną wersję obrazu. Jest wygodny podczas szybkiego uruchamiania aktualnej wersji aplikacji.

Tag `sha-<commit_sha>` pozwala jednoznacznie powiązać obraz z konkretnym commitem w repozytorium. Dzięki temu można sprawdzić, z jakiej wersji kodu powstał dany obraz. Takie tagowanie ułatwia odtwarzalność procesu budowania, analizę błędów oraz ewentualny powrót do wcześniejszej wersji.

Zastosowanie obu tagów łączy wygodę użycia tagu `latest` z jednoznaczną identyfikacją wersji przez hash commita.

Źródło:

```text
https://docs.docker.com/build/ci/github-actions/manage-tags-labels/
https://github.com/docker/metadata-action
```

## Schemat tagowania danych cache

Dane cache są przechowywane w dedykowanym publicznym repozytorium DockerHub:

```text
kanterete/zadanie-2-cache
```

Zastosowano cache typu `registry` oraz tryb `mode=max`.

Dla etapu skanowania cache jest rozdzielony według architektury:

```text
amd64
arm64
```

Natomiast dla finalnego budowania obrazu multi-arch wykorzystywany jest tag:

```text
buildcache
```

Przyjęty schemat tagowania cache:

```text
kanterete/zadanie-2-cache:amd64
kanterete/zadanie-2-cache:arm64
kanterete/zadanie-2-cache:buildcache
```

Rozdzielenie cache według architektury pozwala uniknąć mieszania warstw budowanych dla różnych platform. Z kolei tag `buildcache` służy do zapisu cache z finalnego procesu budowania obrazu multi-arch.

Tryb `mode=max` został użyty, ponieważ pozwala zapisać większy zakres danych cache BuildKit, co zwiększa szansę ponownego wykorzystania warstw w kolejnych uruchomieniach pipeline’u.

Źródło:

```text
https://docs.docker.com/build/ci/github-actions/cache/
```

## Multi-arch build

Do budowania obrazu dla wielu architektur wykorzystano Docker Buildx oraz QEMU.

W workflow ustawiono:

```yaml
platforms: linux/amd64,linux/arm64
```

Dzięki temu publikowany obraz obsługuje dwie wymagane architektury sprzętowe.

Źródło:

```text
https://docs.docker.com/build/building/multi-platform/
```

## Potwierdzenie działania

Workflow został uruchomiony ręcznie za pomocą `workflow_dispatch`.

Do obserwowania działania pipeline’u użyto polecenia:

```bash
gh run watch
```

Uruchomienie zakończyło się statusem:

```text
success
```

Poprawnie wykonane zostały joby:

```text
Build and scan linux/arm64
Build and scan linux/amd64
Publish multi-arch image to GHCR
```

Czas wykonania workflow wyniósł około 1 minutę i 52 sekundy.

Po zakończeniu workflow obraz pojawił się w GitHub Container Registry jako publiczny pakiet:

```text
ghcr.io/kanterete/pawcho-zadanie-2
```

Dostępne tagi obrazu:

```text
latest
sha-98c64107dd8c5b3eadd4e2463f6b7a043b2dc0d2
```

## Podsumowanie

Przygotowany pipeline spełnia wymagania zadania. Obraz kontenera jest budowany dla architektur `linux/amd64` i `linux/arm64`, wykorzystuje cache przechowywany w DockerHub, przechodzi test CVE z użyciem Trivy, a następnie jest publikowany do GitHub Container Registry tylko po pozytywnym zakończeniu skanowania.
