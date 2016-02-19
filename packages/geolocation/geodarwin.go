// +build darwin
package geolocation

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreLocation
#import <Foundation/Foundation.h>
#import <CoreLocation/CoreLocation.h>
@interface Location : NSObject <CLLocationManagerDelegate>
@property (nonatomic, strong)CLLocationManager *manager;
@end
@implementation Location
- (instancetype)init {
    if (self = [super init]) {
        _manager = [[CLLocationManager alloc] init];
        _manager.delegate = self;
        [_manager startUpdatingLocation];
    }
    return self;
}
- (void)locationManager:(CLLocationManager *)manager didUpdateLocations:(NSArray *)locations
{
    NSLog(@"%@", locations);
}
@end
char* getLocation() {
    @autoreleasepool {
        Location *location = [[Location alloc] init];
        while (location.manager.location == nil) {
        }
        NSString *str = [NSString stringWithFormat:@"%f, %f", location.manager.location.coordinate.latitude,
                         location.manager.location.coordinate.longitude];
        return (char*)[str UTF8String];
    }
}
*/
import "C"

import (
	"strings"
	"strconv"
	"errors"
	"fmt"
)

func goString(s *C.char) string {
	return C.GoString(s)
}

func CLLocation() (*coordinates, error) {
	str := goString(C.getLocation())
	sCoords := strings.Split(str, ", ")
	if len(sCoords) != 2 {
		return nil, errors.New("Wrong coordinates")
	}
	fmt.Println("Calling CLLocation()")
	lat, _ := strconv.ParseFloat(sCoords[0], 64)
	lng, _ := strconv.ParseFloat(sCoords[1], 64)

	return &coordinates{
		Latitude:lat,
		Longitude:lng,
	}, nil
}