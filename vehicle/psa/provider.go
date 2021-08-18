package psa

import (
	"strings"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/provider"
)

// Provider is an api.Vehicle implementation for PSA cars
type Provider struct {
	statusG func() (Status, error)
}

// NewProvider creates a new vehicle
func NewProvider(api *API, vid string, cache time.Duration) *Provider {
	impl := &Provider{
		statusG: provider.NewCached[Status](func() (Status, error) {
			return api.Status(vid)
		}, cache).Get,
	}
	return impl
}

var _ api.Battery = (*Provider)(nil)

// SoC implements the api.Vehicle interface
func (v *Provider) SoC() (float64, error) {
	res, err := v.statusG()
	if err == nil {
		for _, e := range res.Energy {
			if e.Type != "Electric" {
				continue
			}

			return float64(e.Level), nil
		}

		err = api.ErrNotAvailable
	}

	return 0, err
}

var _ api.VehicleRange = (*Provider)(nil)

// Range implements the api.VehicleRange interface
func (v *Provider) Range() (int64, error) {
	res, err := v.statusG()
	if err == nil {
		for _, e := range res.Energy {
			if e.Type != "Electric" {
				continue
			}

			return int64(e.Autonomy), nil
		}

		err = api.ErrNotAvailable
	}

	return 0, err
}

var _ api.VehicleOdometer = (*Provider)(nil)

// Odometer implements the api.VehicleOdometer interface
func (v *Provider) Odometer() (float64, error) {
	res, err := v.statusG()

	if err == nil {
		return res.Odometer.Mileage, nil
	}

	return 0, err
}

var _ api.VehicleFinishTimer = (*Provider)(nil)

// FinishTime implements the api.VehicleFinishTimer interface
func (v *Provider) FinishTime() (time.Time, error) {
	res, err := v.statusG()
	if err == nil {
		for _, e := range res.Energy {
			if e.Type != "Electric" {
				continue
			}

			return e.UpdatedAt.Add(e.Charging.RemainingTime.Duration), nil
		}

		err = api.ErrNotAvailable
	}

	return time.Time{}, err
}

var _ api.ChargeState = (*Provider)(nil)

// Status implements the api.ChargeState interface
func (v *Provider) Status() (api.ChargeStatus, error) {
	res, err := v.statusG()
	if err == nil {
		for _, e := range res.Energy {
			if e.Type != "Electric" {
				continue
			}

			status := api.StatusA

			if e.Charging.Plugged {
				status = api.StatusB

				if strings.ToLower(e.Charging.Status) == "inprogress" {
					status = api.StatusC
				}
			}

			return status, nil
		}

		err = api.ErrNotAvailable
	}

	return api.StatusNone, err
}

var _ api.VehicleClimater = (*Provider)(nil)

// Climater implements the api.VehicleClimater interface
func (v *Provider) Climater() (active bool, outsideTemp float64, targetTemp float64, err error) {
	res, err := v.statusG()
	if err == nil {
		active := strings.ToLower(res.Preconditionning.AirConditioning.Status) != "disabled"
		return active, 20, 20, nil
	}

	return active, outsideTemp, targetTemp, err
}
