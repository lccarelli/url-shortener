package service

import (
	"URLShortest/config"
	"URLShortest/model"
	"URLShortest/repository"
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"hash/crc32"
	"strconv"
	"time"
)

// ShortenerService es el servicio principal para acortar URLs.
type ShortenerService struct {
	repo *repository.RedisRepository
	ttl  time.Duration
}

// NewShortenerService crea una nueva instancia del servicio.
func NewShortenerService(repo *repository.RedisRepository, ttl time.Duration) *ShortenerService {
	return &ShortenerService{
		repo: repo,
		ttl:  ttl,
	}
}

// generateKey genera un hash base36 a partir de la URL original.
func generateKey(url string) string {
	hash := crc32.ChecksumIEEE([]byte(url))
	return strconv.FormatUint(uint64(hash), 36)
}

// ShortenURL acorta una URL y la guarda como una entrada única siempre.
func (s *ShortenerService) ShortenURL(ctx context.Context, req model.ShortenRequest) (*model.ShortenResponse, error) {
	tr := otel.Tracer("shortener-service")
	ctx, span := tr.Start(ctx, "ShortenURL")
	defer span.End()

	log := config.Log.WithField("original_url", req.URL)

	if req.URL == "" {
		log.Warn("URL is empty")
		return nil, ErrInvalidURL
	}

	// Generar clave única cada vez
	uniqueComponent := strconv.FormatInt(time.Now().UnixNano(), 36)
	key := generateKey(req.URL + "-" + uniqueComponent)

	log = log.WithField("short_key", key)
	log.Info("Generated unique short key for URL")

	// Guardar en Redis
	if err := s.repo.Set(key, req.URL, s.ttl); err != nil {
		log.WithError(err).Error("Failed to write short URL to Redis")
		return nil, err
	}

	// Operaciones async no bloqueantes
	go func() {
		if err := s.repo.AddToGlobalSet(key); err != nil {
			log.WithError(err).Warn("Failed to add to global set")
		}
		if err := s.repo.InitRanking(key); err != nil {
			log.WithError(err).Warn("Failed to init ranking")
		}
		if err := s.repo.IncrementGlobalStat("created"); err != nil {
			log.WithError(err).Warn("Failed to increment created counter")
		}
	}()

	log.Info("Shorten successful")
	return &model.ShortenResponse{
		Short: key,
		Url:   "http://localhost:8080/redirect/" + key,
	}, nil
}

// Errores específicos del dominio
var (
	ErrInvalidURL    = errors.New("invalid URL")
	ErrHashCollision = errors.New("hash collision detected")
)
