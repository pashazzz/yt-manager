package handlers

import (
	"context"
	"log"
	"time"

	"github.com/pavlo/yt-manager/internal/providers"
	"github.com/pavlo/yt-manager/internal/repository"
)

// BackfillThumbnails в фоне находит эпизоды без thumbnailUrl и пытается
// получить превью через соответствующего провайдера. Используется на старте
// сервера для миграции данных, добавленных до появления поля thumbnailUrl.
//
// Стратегия:
//   - Группируем эпизоды по providerName, далее обрабатываем по одному:
//     для каждого вызываем provider.Fetch с одиночным URL и берём BestThumbnail.
//   - Между запросами делаем небольшую паузу, чтобы не упереться в rate-limit yt-dlp.
//   - Ошибки тихо игнорируются — следующий запуск попробует снова.
func BackfillThumbnails(eps *repository.EpisodeRepo, registry *providers.Registry) {
	missing, err := eps.FindMissingThumbnails()
	if err != nil {
		log.Printf("backfill: failed to query episodes: %v", err)
		return
	}
	if len(missing) == 0 {
		return
	}
	log.Printf("backfill: %d episode(s) without thumbnailUrl, processing in background", len(missing))

	updated := 0
	for _, ep := range missing {
		if ep.Provider == "" || ep.VideoID == "" {
			continue
		}
		p := registry.ByName(ep.Provider)
		if p == nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		info, ferr := p.Fetch(ctx, p.VideoURL(ep.VideoID))
		cancel()
		if ferr != nil || info == nil || len(info.Entries) == 0 {
			continue
		}
		thumb := info.Entries[0].BestThumbnail()
		if thumb == "" {
			continue
		}
		if uerr := eps.UpdateThumbnail(ep.ID, thumb); uerr != nil {
			log.Printf("backfill: update %s failed: %v", ep.ID, uerr)
			continue
		}
		updated++
		// Дросселирование, чтобы не дудосить провайдер.
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("backfill: completed, updated %d episode(s)", updated)
}
